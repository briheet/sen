package graph

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

const (
	maxPooledRenderNodes  = 4096
	maxPooledRenderPixels = 2_000_000
)

type renderRequest struct {
	nodes                  []node
	edges                  []edgeModel
	buffer                 *renderNodeBuffer
	camera                 camera
	sequence               uint64
	imageID                uint32
	previousImageID        uint32
	dragging, selected     int
	hovered, width, height int
	originX, originY       int
	cellWidth, cellHeight  int
	nodeRadius             float64
	obscured               bool
	quiet                  byte
	done                   chan renderResult
}

type renderResult struct {
	frame   frameOutput
	err     error
	dropped bool
}

// frameOutput carries terminal bytes and the state committed after they print.
type frameOutput struct {
	output   string
	owner    string
	buffer   *renderNodeBuffer
	sequence uint64
	imageID  uint32
}

func (m frameOutput) String() string { return m.output }

type renderReadyMsg struct {
	frame frameOutput
}

// DebugSummary avoids copying encoded image bytes into diagnostic logs.
func (m renderReadyMsg) DebugSummary() string {
	return fmt.Sprintf("graph[%s] render ready sequence=%d image_id=%d",
		m.frame.owner, m.frame.sequence, m.frame.imageID)
}

// renderer serializes image encoding and retains its working memory.
type renderer struct {
	dump      io.Writer
	owner     string
	palette   graphPalette
	nodeClass []uint8
	requests  chan renderRequest
	latest    atomic.Uint64
	start     sync.Once

	output bytes.Buffer
	chunks kittyChunkWriter
	png    encoderBufferPool
}

type encoderBufferPool struct{ buffer *png.EncoderBuffer }

type renderNodeBuffer struct {
	nodes []node
	edges []edgeModel
}

type renderSurface struct {
	canvas *image.Paletted
	blur   []uint16
}

var renderNodes = sync.Pool{New: func() any { return new(renderNodeBuffer) }}
var renderSurfaces = sync.Pool{New: func() any { return new(renderSurface) }}

func (p *encoderBufferPool) Get() *png.EncoderBuffer {
	buffer := p.buffer
	p.buffer = nil
	return buffer
}

func (p *encoderBufferPool) Put(buffer *png.EncoderBuffer) {
	p.buffer = buffer
}

func newRenderer(owner string, dump io.Writer) *renderer {
	return &renderer{
		owner:    owner,
		dump:     dump,
		palette:  newGraphPalette(styles.Zakura),
		requests: make(chan renderRequest, 1),
	}
}

func (r *renderer) submit(request renderRequest) {
	r.start.Do(func() { go r.run() })
	r.latest.Store(request.sequence)
	select {
	case r.requests <- request:
	default:
		select {
		case stale := <-r.requests:
			releaseRenderNodes(stale.buffer)
			stale.done <- renderResult{dropped: true}
		default:
		}
		r.requests <- request
	}
}

// cancel makes queued or in-flight frames stale before a graph is hidden.
func (r *renderer) cancel(sequence uint64) {
	r.latest.Store(sequence)
}

func (r *renderer) run() {
	for request := range r.requests {
		err := r.render(&request)
		result := renderResult{err: err}
		if request.sequence != r.latest.Load() {
			result = renderResult{dropped: true}
		} else if err == nil {
			result.frame = frameOutput{
				output: r.output.String(), owner: r.owner, buffer: request.buffer,
				sequence: request.sequence, imageID: request.imageID,
			}
			r.trace("frame ready image_id=%d previous_id=%d bytes=%d origin=%d,%d size=%dx%d",
				request.imageID, request.previousImageID, r.output.Len(),
				request.originX, request.originY, request.width, request.height)
		}
		request.done <- result
		if result.dropped || result.err != nil {
			releaseRenderNodes(request.buffer)
		}
	}
}

func snapshotGraph(nodes []node, edges []edgeModel) *renderNodeBuffer {
	buffer := renderNodes.Get().(*renderNodeBuffer)
	if cap(buffer.nodes) < len(nodes) {
		buffer.nodes = make([]node, len(nodes))
	} else {
		buffer.nodes = buffer.nodes[:len(nodes)]
	}
	if cap(buffer.edges) < len(edges) {
		buffer.edges = make([]edgeModel, len(edges))
	} else {
		buffer.edges = buffer.edges[:len(edges)]
	}
	copy(buffer.nodes, nodes)
	copy(buffer.edges, edges)
	return buffer
}

func releaseRenderNodes(buffer *renderNodeBuffer) {
	if buffer == nil || cap(buffer.nodes) > maxPooledRenderNodes || cap(buffer.edges) > maxPooledRenderNodes {
		return
	}
	clear(buffer.nodes)
	clear(buffer.edges)
	buffer.nodes = buffer.nodes[:0]
	buffer.edges = buffer.edges[:0]
	renderNodes.Put(buffer)
}

func (r *renderer) render(request *renderRequest) error {
	r.output.Reset()
	bounds := image.Rect(0, 0, request.width*request.cellWidth, request.height*request.cellHeight)
	surface := acquireRenderSurface(bounds, r.palette.colors)
	defer releaseRenderSurface(surface)
	r.renderImage(request, surface.canvas)
	if request.obscured {
		softenRenderSurface(surface, &r.palette)
	}
	options := &kitty.Options{
		Action:       kitty.Transmit,
		ID:           int(request.imageID),
		Format:       kitty.PNG,
		Transmission: kitty.Direct,
		Quiet:        request.quiet,
	}
	if err := r.encodeImage(surface.canvas, options); err != nil {
		r.trace("upload failed error=%q", err)
		return err
	}
	// Upload first, then atomically swap physical placements at the viewport.
	placement := &kitty.Options{
		Action:          kitty.Put,
		ID:              int(request.imageID),
		Quiet:           request.quiet,
		Columns:         request.width,
		Rows:            request.height,
		DoNotMoveCursor: true,
	}
	r.output.WriteString(ansi.SetModeSynchronizedOutput)
	r.output.WriteString(ansi.SaveCursor)
	r.output.WriteString(ansi.CursorPosition(request.originX+1, request.originY+1))
	placementOptions := append(placement.Options(), "z=-1")
	r.output.WriteString(ansi.KittyGraphics(nil, placementOptions...))
	if request.previousImageID != 0 {
		remove := &kitty.Options{
			Action:          kitty.Delete,
			ID:              int(request.previousImageID),
			Quiet:           request.quiet,
			Delete:          kitty.DeleteID,
			DeleteResources: true,
		}
		r.output.WriteString(ansi.KittyGraphics(nil, remove.Options()...))
	}
	r.output.WriteString(ansi.RestoreCursor)
	r.output.WriteString(ansi.ResetModeSynchronizedOutput)
	return nil
}

func (r *renderer) encodeImage(source image.Image, options *kitty.Options) error {
	r.chunks.reset(&r.output, options)
	base64Writer := base64.NewEncoder(base64.StdEncoding, &r.chunks)
	encoder := png.Encoder{CompressionLevel: png.BestSpeed, BufferPool: &r.png}
	if err := encoder.Encode(base64Writer, source); err != nil {
		return err
	}
	if err := base64Writer.Close(); err != nil {
		return err
	}
	r.chunks.close()
	return nil
}

func acquireRenderSurface(bounds image.Rectangle, palette color.Palette) *renderSurface {
	surface := renderSurfaces.Get().(*renderSurface)
	if surface.canvas == nil || surface.canvas.Bounds() != bounds {
		surface.canvas = image.NewPaletted(bounds, palette)
	} else {
		surface.canvas.Palette = palette
	}
	return surface
}

func releaseRenderSurface(surface *renderSurface) {
	// Avoid retaining an unusually large terminal surface through the pool.
	if surface == nil || surface.canvas == nil || len(surface.canvas.Pix) > maxPooledRenderPixels {
		return
	}
	renderSurfaces.Put(surface)
}

// softenRenderSurface applies a separable box blur and lowers graph contrast.
func softenRenderSurface(surface *renderSurface, palette *graphPalette) {
	const radius = 6
	canvas := surface.canvas
	width, height := canvas.Bounds().Dx(), canvas.Bounds().Dy()
	if width == 0 || height == 0 {
		return
	}
	if cap(surface.blur) < len(canvas.Pix) {
		surface.blur = make([]uint16, len(canvas.Pix))
	} else {
		surface.blur = surface.blur[:len(canvas.Pix)]
	}
	level := func(index uint8) int {
		if index == 0 {
			return 0
		}
		return (int(index)-1)%alphaLevels + 1
	}
	for y := range height {
		row := y * canvas.Stride
		sum := 0
		for x := 0; x <= min(radius, width-1); x++ {
			sum += level(canvas.Pix[row+x])
		}
		for x := range width {
			surface.blur[row+x] = uint16(sum)
			if next := x + radius + 1; next < width {
				sum += level(canvas.Pix[row+next])
			}
			if previous := x - radius; previous >= 0 {
				sum -= level(canvas.Pix[row+previous])
			}
		}
	}
	for x := range width {
		sum := 0
		for y := 0; y <= min(radius, height-1); y++ {
			sum += int(surface.blur[y*canvas.Stride+x])
		}
		for y := range height {
			top, bottom := max(0, y-radius), min(height-1, y+radius)
			left, right := max(0, x-radius), min(width-1, x+radius)
			blurred := sum / ((bottom - top + 1) * (right - left + 1)) * 2 / 5
			if blurred == 0 {
				canvas.Pix[y*canvas.Stride+x] = 0
			} else {
				canvas.Pix[y*canvas.Stride+x] = palette.idle[min(alphaLevels, blurred)]
			}
			if next := y + radius + 1; next < height {
				sum += int(surface.blur[next*canvas.Stride+x])
			}
			if previous := y - radius; previous >= 0 {
				sum -= int(surface.blur[previous*canvas.Stride+x])
			}
		}
	}
}

func (r *renderer) trace(format string, values ...any) {
	if r.dump != nil {
		_, _ = fmt.Fprintf(r.dump, "graph[%s] "+format+"\n", append([]any{r.owner}, values...)...)
	}
}

type renderFailedMsg struct {
	owner string
	err   error
}

func renderCommand(done <-chan renderResult, owner string) tea.Cmd {
	return func() tea.Msg {
		result := <-done
		switch {
		case result.dropped:
			return nil
		case result.err != nil:
			return renderFailedMsg{owner: owner, err: result.err}
		default:
			return renderReadyMsg{frame: result.frame}
		}
	}
}

func deleteImagesCommand(ids [2]uint32, quiet byte) tea.Cmd {
	var output strings.Builder
	output.WriteString(ansi.SetModeSynchronizedOutput)
	for _, id := range ids {
		remove := &kitty.Options{
			Action: kitty.Delete, ID: int(id), Quiet: quiet,
			Delete: kitty.DeleteID, DeleteResources: true,
		}
		output.WriteString(ansi.KittyGraphics(nil, remove.Options()...))
	}
	output.WriteString(ansi.ResetModeSynchronizedOutput)
	return tea.Raw(output.String())
}
