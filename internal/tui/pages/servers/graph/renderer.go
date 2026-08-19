package graph

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"sync"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"golang.org/x/image/font"
)

const maxPooledRenderNodes = 4096

type renderRequest struct {
	face                    font.Face
	nodes                   []node
	sequence                uint64
	imageID                 uint32
	dragging, width, height int
	cellWidth, cellHeight   int
	nodeRadius              float64
	quiet                   byte
	done                    chan renderResult
}

type renderResult struct {
	output  string
	err     error
	dropped bool
}

// renderer serializes image encoding and retains its working memory.
type renderer struct {
	dump     io.Writer
	owner    string
	edges    []edgeModel
	palette  graphPalette
	canvas   *image.Paletted
	requests chan renderRequest
	latest   atomic.Uint64
	start    sync.Once

	output  bytes.Buffer
	encoded bytes.Buffer
	payload bytes.Buffer
	png     encoderBufferPool
}

type encoderBufferPool struct{ buffer *png.EncoderBuffer }

var renderNodes sync.Pool

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
			releaseRenderNodes(stale.nodes)
			stale.done <- renderResult{dropped: true}
		default:
		}
		r.requests <- request
	}
}

func (r *renderer) run() {
	for request := range r.requests {
		err := r.render(request)
		result := renderResult{err: err}
		if request.sequence != r.latest.Load() {
			result = renderResult{dropped: true}
		} else if err == nil {
			result.output = r.output.String()
			r.trace("upload queued image_id=%d bytes=%d width=%d height=%d",
				request.imageID, r.output.Len(), request.width, request.height)
		}
		request.done <- result
		releaseRenderNodes(request.nodes)
	}
}

func snapshotNodes(source []node) []node {
	nodes, _ := renderNodes.Get().([]node)
	if cap(nodes) < len(source) {
		nodes = make([]node, len(source))
	} else {
		nodes = nodes[:len(source)]
	}
	copy(nodes, source)
	return nodes
}

func releaseRenderNodes(nodes []node) {
	if cap(nodes) > maxPooledRenderNodes {
		return
	}
	clear(nodes)
	renderNodes.Put(nodes[:0])
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
	placement := &kitty.Options{
		Action:           kitty.Put,
		ID:               int(request.imageID),
		Quiet:            request.quiet,
		Columns:          request.width,
		Rows:             request.height,
		VirtualPlacement: true,
		DoNotMoveCursor:  true,
	}
	_, err := r.output.WriteString(ansi.KittyGraphics(nil, placement.Options()...))
	if err != nil {
		r.trace("upload failed error=%q", err)
		return err
	}
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

// InvalidateView asks the root model to replace the graph with its error state.
func (renderFailedMsg) InvalidateView() {}

func renderCommand(done <-chan renderResult, owner string) tea.Cmd {
	return func() tea.Msg {
		result := <-done
		switch {
		case result.dropped:
			return nil
		case result.err != nil:
			return renderFailedMsg{owner: owner, err: result.err}
		default:
			return tea.Raw(result.output)()
		}
	}
}
