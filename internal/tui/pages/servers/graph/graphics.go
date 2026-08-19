package graph

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"golang.org/x/sys/unix"
)

const (
	maxPlaceholderSpan  = 283
	minimumRasterHeight = 12
)

var errViewportTooLarge = errors.New("graph viewport exceeds Kitty placeholder limits")

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
	common := width
	for value := height; value != 0; {
		common, value = value, common%value
	}
	width, height = width/common, height/common
	scale := max(1, (minimumRasterHeight+height-1)/height)
	return width * scale, height * scale
}

func (m *Model) upload() tea.Cmd {
	if !m.graphics || len(m.nodes) == 0 || m.width == 0 || m.height == 0 {
		return nil
	}
	if m.width > maxPlaceholderSpan || m.height > maxPlaceholderSpan {
		m.renderErr = errViewportTooLarge
		m.trace("upload rejected error=%q", m.renderErr)
		return nil
	}
	quiet := byte(2)
	if m.dump != nil {
		quiet = 1 // Keep terminal errors visible while debugging.
	}
	m.renderSequence++
	done := make(chan renderResult, 1)
	nodes := snapshotNodes(m.nodes)
	m.renderer.submit(renderRequest{
		face:       m.labelFace,
		nodes:      nodes,
		sequence:   m.renderSequence,
		imageID:    m.imageID,
		dragging:   m.dragging,
		width:      m.width,
		height:     m.height,
		cellWidth:  m.cellWidth,
		cellHeight: m.cellHeight,
		nodeRadius: m.nodeRadius,
		quiet:      quiet,
		done:       done,
	})
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

func placeholders(id uint32, width, height int) (string, error) {
	if width > maxPlaceholderSpan || height > maxPlaceholderSpan {
		return "", errViewportTooLarge
	}
	colour := ansi.RGBColor{R: byte(id >> 16), G: byte(id >> 8), B: byte(id)}
	foreground := ansi.Style{}.ForegroundColor(colour).String()
	resetForeground := ansi.Style{}.ForegroundColor(nil).String()
	var output strings.Builder
	for row := range height {
		output.WriteString(foreground)
		for column := range width {
			output.WriteRune(kitty.Placeholder)
			output.WriteRune(kitty.Diacritic(row))
			output.WriteRune(kitty.Diacritic(column))
		}
		output.WriteString(resetForeground)
		if row < height-1 {
			output.WriteByte('\n')
		}
	}
	return output.String(), nil
}

func centeredMessage(width, height int, message string) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(message)
}
