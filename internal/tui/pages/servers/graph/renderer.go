package graph

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"golang.org/x/image/vector"
)

const maxPooledRenderNodes = 4096

type renderRequest struct {
	nodes                   []node
	buffer                  *renderNodeBuffer
	sequence                uint64
	imageID                 uint32
	previousImageID         uint32
	dragging, width, height int
	originX, originY        int
	cellWidth, cellHeight   int
	nodeRadius              float64
	quiet                   byte
	done                    chan renderResult
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
	dump       io.Writer
	owner      string
	edges      []edgeModel
	palette    graphPalette
	canvas     *image.Paletted
	mask       *image.Alpha
	rasterizer vector.Rasterizer
	requests   chan renderRequest
	latest     atomic.Uint64
	start      sync.Once

	output  bytes.Buffer
	encoded bytes.Buffer
	payload bytes.Buffer
	png     encoderBufferPool
}

type encoderBufferPool struct{ buffer *png.EncoderBuffer }

type renderNodeBuffer struct{ nodes []node }

var renderNodes = sync.Pool{New: func() any { return new(renderNodeBuffer) }}

func (p *encoderBufferPool) Get() *png.EncoderBuffer {
	buffer := p.buffer
	p.buffer = nil
	return buffer
}

func (p *encoderBufferPool) Put(buffer *png.EncoderBuffer) {
	p.buffer = buffer
}

func newRenderer(owner string, edges []edgeModel, dump io.Writer) *renderer {
	return &renderer{
		owner:    owner,
		dump:     dump,
		edges:    append([]edgeModel(nil), edges...),
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
		err := r.render(request)
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

func snapshotNodes(source []node) *renderNodeBuffer {
	buffer := renderNodes.Get().(*renderNodeBuffer)
	if cap(buffer.nodes) < len(source) {
		buffer.nodes = make([]node, len(source))
	} else {
		buffer.nodes = buffer.nodes[:len(source)]
	}
	copy(buffer.nodes, source)
	return buffer
}

func releaseRenderNodes(buffer *renderNodeBuffer) {
	if buffer == nil || cap(buffer.nodes) > maxPooledRenderNodes {
		return
	}
	clear(buffer.nodes)
	buffer.nodes = buffer.nodes[:0]
	renderNodes.Put(buffer)
}

func (r *renderer) render(request renderRequest) error {
	r.output.Reset()
	image := r.renderImage(request)
	options := &kitty.Options{
		Action:       kitty.Transmit,
		ID:           int(request.imageID),
		Format:       kitty.PNG,
		Transmission: kitty.Direct,
		Quiet:        request.quiet,
	}
	if err := r.encodeImage(image, options); err != nil {
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
	r.encoded.Reset()
	r.payload.Reset()
	encoder := png.Encoder{CompressionLevel: png.BestSpeed, BufferPool: &r.png}
	if err := encoder.Encode(&r.encoded, source); err != nil {
		return err
	}
	base64Writer := base64.NewEncoder(base64.StdEncoding, &r.payload)
	if _, err := r.encoded.WriteTo(base64Writer); err != nil {
		return err
	}
	if err := base64Writer.Close(); err != nil {
		return err
	}
	writeChunks(&r.output, r.payload.Bytes(), options)
	return nil
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
