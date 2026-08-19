package graph

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	buffer := snapshotNodes(m.nodes)
	m.renderer.submit(renderRequest{
		nodes:           buffer.nodes,
		buffer:          buffer,
		sequence:        m.renderSequence,
		imageID:         imageID,
		previousImageID: m.frontImageID,
		dragging:        m.dragging,
		width:           m.width,
		height:          m.height,
		originX:         m.originX,
		originY:         m.originY,
		cellWidth:       m.cellWidth,
		cellHeight:      m.cellHeight,
		nodeRadius:      m.nodeRadius,
		quiet:           quiet,
		done:            done,
	})
	m.renderPending = true
	m.renderErr = nil
	return renderCommand(done, m.owner)
}

func writeChunks(output *bytes.Buffer, payload []byte, options *kitty.Options) {
	for offset := 0; offset < len(payload); offset += kitty.MaxChunkSize {
		end := min(offset+kitty.MaxChunkSize, len(payload))
		var values []string
		if offset == 0 {
			values = options.Options()
		} else if options.Quiet > 0 {
			values = []string{"q=" + strconv.Itoa(int(options.Quiet))}
		}
		if len(payload) > kitty.MaxChunkSize {
			continuation := "m=1"
			if end == len(payload) {
				continuation = "m=0"
			}
			values = append(values, continuation)
		}
		output.WriteString(ansi.KittyGraphics(payload[offset:end], values...))
	}
}

func centeredMessage(width, height int, message string) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(message)
}
