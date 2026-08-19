package graph

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

const (
	maxPlaceholderSpan    = 283
	maxPooledBufferMemory = 1 << 20
)

var (
	encodeBuffers sync.Pool
	pngBuffers    encoderBufferPool
)

type encoderBufferPool struct{ buffers sync.Pool }

func (p *encoderBufferPool) Get() *png.EncoderBuffer {
	buffer, _ := p.buffers.Get().(*png.EncoderBuffer)
	return buffer
}

func (p *encoderBufferPool) Put(buffer *png.EncoderBuffer) {
	p.buffers.Put(buffer)
}

var errViewportTooLarge = errors.New("graph viewport exceeds Kitty placeholder limits")

func supportsGraphics() bool {
	terminal := strings.ToLower(os.Getenv("TERM"))
	program := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	return strings.Contains(terminal, "ghostty") || strings.Contains(terminal, "kitty") ||
		strings.Contains(program, "ghostty") || strings.Contains(program, "kitty") ||
		os.Getenv("KITTY_WINDOW_ID") != ""
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
	output := acquireBuffer()
	defer releaseBuffer(output)
	err := encodeImage(output, m.renderImage(), &kitty.Options{
		Action:       kitty.Transmit,
		ID:           int(m.imageID),
		Format:       kitty.PNG,
		Transmission: kitty.Direct,
		Quiet:        quiet,
	})
	if err == nil {
		// A separate virtual placement keeps image transfer independent of layout.
		placement := &kitty.Options{
			Action:           kitty.Put,
			ID:               int(m.imageID),
			Quiet:            quiet,
			Columns:          m.width,
			Rows:             m.height,
			VirtualPlacement: true,
			DoNotMoveCursor:  true,
		}
		_, err = output.WriteString(ansi.KittyGraphics(nil, placement.Options()...))
	}
	if err != nil {
		m.renderErr = err
		m.trace("upload failed error=%q", err)
		return nil
	}
	m.renderErr = nil
	m.trace("upload queued image_id=%d bytes=%d width=%d height=%d", m.imageID, output.Len(), m.width, m.height)
	return tea.Raw(output.String())
}

// encodeImage reuses PNG and byte buffers across animation frames.
func encodeImage(output *bytes.Buffer, image *image.RGBA, options *kitty.Options) error {
	encoded := acquireBuffer()
	payload := acquireBuffer()
	defer releaseBuffer(encoded)
	defer releaseBuffer(payload)

	encoder := png.Encoder{CompressionLevel: png.BestSpeed, BufferPool: &pngBuffers}
	if err := encoder.Encode(encoded, image); err != nil {
		return err
	}
	base64Writer := base64.NewEncoder(base64.StdEncoding, payload)
	if _, err := encoded.WriteTo(base64Writer); err != nil {
		return err
	}
	if err := base64Writer.Close(); err != nil {
		return err
	}
	writeChunks(output, payload.Bytes(), options)
	return nil
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

func acquireBuffer() *bytes.Buffer {
	buffer, _ := encodeBuffers.Get().(*bytes.Buffer)
	if buffer == nil {
		return new(bytes.Buffer)
	}
	return buffer
}

func releaseBuffer(buffer *bytes.Buffer) {
	if buffer.Cap() > maxPooledBufferMemory {
		return
	}
	buffer.Reset()
	encodeBuffers.Put(buffer)
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
