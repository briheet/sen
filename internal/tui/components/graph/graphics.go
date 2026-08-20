package graph

import (
	"bytes"
	"math"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi/kitty"
	"golang.org/x/sys/unix"
)

const maximumRasterHeight = 20

func supportsGraphics() bool {
	terminal := strings.ToLower(os.Getenv("TERM"))
	program := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	return strings.Contains(terminal, "ghostty") || strings.Contains(terminal, "kitty") ||
		strings.Contains(program, "ghostty") || strings.Contains(program, "kitty") ||
		os.Getenv("KITTY_WINDOW_ID") != ""
}

func terminalCellSize() (int, int) {
	for _, descriptor := range []uintptr{os.Stdout.Fd(), os.Stdin.Fd()} {
		size, err := unix.IoctlGetWinsize(int(descriptor), unix.TIOCGWINSZ)
		if err == nil && size.Col > 0 && size.Row > 0 && size.Xpixel > 0 && size.Ypixel > 0 {
			width := (int(size.Xpixel) + int(size.Col)/2) / int(size.Col)
			height := (int(size.Ypixel) + int(size.Row)/2) / int(size.Row)
			return rasterCellSize(max(1, width), max(1, height))
		}
	}
	return fallbackCellWidth, fallbackCellHeight
}

func rasterCellSize(width, height int) (int, int) {
	if height <= maximumRasterHeight {
		return width, height
	}
	// Kitty scales the image to terminal cells, so cap work while preserving aspect.
	return max(1, (width*maximumRasterHeight+height/2)/height), maximumRasterHeight
}

func (m Model) quiet() byte {
	if m.dump != nil {
		return 1 // Keep terminal protocol errors visible while debugging.
	}
	return 2
}

func (m *Model) upload() tea.Cmd {
	if !m.graphics || !m.visible || m.renderPending || len(m.nodes) == 0 || m.width == 0 || m.height == 0 {
		return nil
	}
	quiet := m.quiet()
	m.renderSequence++
	imageID := m.imageIDs[0]
	if m.frontImageID == imageID {
		imageID = m.imageIDs[1]
	}
	done := make(chan renderResult, 1)
	// Keep the committed label position on the node; renderer buffers are pooled.
	for index := range m.nodes {
		m.nodes[index].rendered = m.nodes[index].position
	}
	m.renderedCamera = m.camera
	m.renderedDrag = m.dragging
	m.renderedSelect = m.selected
	m.renderedHover = m.hovered
	m.renderHash = m.visualHash()
	buffer := snapshotGraph(m.nodes, m.edges)
	m.renderer.submit(renderRequest{
		nodes:           buffer.nodes,
		edges:           buffer.edges,
		buffer:          buffer,
		sequence:        m.renderSequence,
		imageID:         imageID,
		previousImageID: m.frontImageID,
		dragging:        m.dragging,
		selected:        m.selected,
		hovered:         m.hovered,
		width:           m.width,
		height:          m.height,
		originX:         m.originX,
		originY:         m.originY,
		cellWidth:       m.cellWidth,
		cellHeight:      m.cellHeight,
		nodeRadius:      m.nodeRadius,
		camera:          m.camera,
		obscured:        m.obscured,
		quiet:           quiet,
		done:            done,
	})
	m.renderPending = true
	m.dirty = false
	m.renderErr = nil
	return renderCommand(done, m.owner)
}

// visualHash quantizes positions to quarter pixels before comparing frames.
func (m *Model) visualHash() uint64 {
	const (
		offset = uint64(14695981039346656037)
		prime  = uint64(1099511628211)
	)
	hash := offset
	mix := func(value uint64) { hash = (hash ^ value) * prime }
	obscured := 0
	if m.obscured {
		obscured = 1
	}
	for _, value := range [...]int{m.width, m.height, m.cellWidth, m.cellHeight, m.dragging, m.selected, m.hovered, obscured} {
		mix(uint64(int64(value)))
	}
	for index := range m.nodes {
		node := &m.nodes[index]
		position := m.camera.worldToScreen(node.position)
		mix(uint64(int64(math.Round(position.x * 4))))
		mix(uint64(int64(math.Round(position.y * 4))))
	}
	return hash
}

// kittyChunkWriter streams base64 output without retaining a second payload.
type kittyChunkWriter struct {
	output  *bytes.Buffer
	options *kitty.Options
	chunk   [kitty.MaxChunkSize]byte
	length  int
	wrote   bool
}

func (w *kittyChunkWriter) reset(output *bytes.Buffer, options *kitty.Options) {
	w.output, w.options = output, options
	w.length = 0
	w.wrote = false
}

func (w *kittyChunkWriter) Write(payload []byte) (int, error) {
	written := len(payload)
	for len(payload) > 0 {
		if w.length == len(w.chunk) {
			w.flush(true)
		}
		count := copy(w.chunk[w.length:], payload)
		w.length += count
		payload = payload[count:]
	}
	return written, nil
}

func (w *kittyChunkWriter) close() {
	if w.length > 0 {
		w.flush(false)
	}
}

func (w *kittyChunkWriter) flush(more bool) {
	var values []string
	if !w.wrote {
		values = w.options.Options()
	} else if w.options.Quiet > 0 {
		values = []string{"q=" + strconv.Itoa(int(w.options.Quiet))}
	}
	if more {
		values = append(values, "m=1")
	} else if w.wrote {
		values = append(values, "m=0")
	}
	writeKittyCommand(w.output, w.chunk[:w.length], values)
	w.length = 0
	w.wrote = true
}

func writeKittyCommand(output *bytes.Buffer, payload []byte, options []string) {
	output.WriteString("\x1b_G")
	for index, option := range options {
		if index > 0 {
			output.WriteByte(',')
		}
		output.WriteString(option)
	}
	if len(payload) > 0 {
		output.WriteByte(';')
		_, _ = output.Write(payload)
	}
	output.WriteString("\x1b\\")
}

func centeredMessage(width, height int, message string) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(message)
}
