package graph

import (
	"image"
	"image/color"
	"math"

	"github.com/briheet/sen/internal/tui/styles"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	pixelsPerCellX = 4
	pixelsPerCellY = 8
	nodeRadius     = 4
	labelGap       = 3
	glyphWidth     = 7
	graphInsetX    = 1.0
	graphInsetY    = 1.0
)

// View returns placeholders whose pixels are supplied through Kitty graphics.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if len(m.nodes) == 0 {
		return centeredMessage(m.width, m.height, "No project functions found.")
	}
	if !m.graphics {
		return centeredMessage(m.width, m.height, "Pixel graph requires Ghostty or Kitty.")
	}
	if !m.ready {
		return centeredMessage(m.width, m.height, "")
	}
	if m.renderErr != nil {
		return centeredMessage(m.width, m.height, "Unable to render graph.")
	}
	if m.placeholder == "" {
		return centeredMessage(m.width, m.height, "Graph viewport is too large.")
	}
	return m.placeholder
}

func (m *Model) renderImage() *image.RGBA {
	bounds := image.Rect(0, 0, m.width*pixelsPerCellX, m.height*pixelsPerCellY)
	if m.canvas == nil || m.canvas.Bounds() != bounds {
		m.canvas = image.NewRGBA(bounds)
	} else {
		clear(m.canvas.Pix)
	}
	canvas := m.canvas
	theme := styles.Zakura
	for _, edge := range m.edges {
		from, to := edgePoints(m.nodes[edge.from], m.nodes[edge.to])
		edge.draw(canvas, from, to, theme)
	}
	drawer := font.Drawer{
		Dst:  canvas,
		Src:  image.NewUniform(theme.Text),
		Face: basicfont.Face7x13,
	}
	for index, node := range m.nodes {
		position := pixelPoint(node.position)
		colour := theme.NodeActive
		if index == m.dragging {
			colour = theme.NodeHot
		}
		drawCircle(canvas, position, nodeRadius, colour, 1)
		drawer.Dot = labelPoint(canvas.Bounds(), position, node.label)
		drawer.DrawString(node.label)
	}
	return canvas
}

func (m edgeModel) draw(canvas *image.RGBA, from, to point, theme styles.Theme) {
	drawLine(canvas, from, to, theme.NodeIdle)
}

// edgePoints connects circle boundaries; labels are a separate render layer.
func edgePoints(from, to node) (point, point) {
	start := pixelPoint(from.position)
	end := pixelPoint(to.position)
	delta := point{x: end.x - start.x, y: end.y - start.y}
	length := distance(start, end)
	if length == 0 {
		return start, end
	}
	offset := min(float64(nodeRadius-1), length/2)
	start.x += delta.x / length * offset
	start.y += delta.y / length * offset
	end.x -= delta.x / length * offset
	end.y -= delta.y / length * offset
	return start, end
}

func labelPoint(bounds image.Rectangle, position point, label string) fixed.Point26_6 {
	textWidth := glyphWidth * len(label)
	x := int(math.Round(position.x)) - textWidth/2
	y := int(math.Round(position.y)) - nodeRadius - labelGap
	ascent := basicfont.Face7x13.Metrics().Ascent.Ceil()
	if y < ascent {
		x = int(math.Round(position.x)) + nodeRadius + labelGap
		y = int(math.Round(position.y)) + ascent/2
	}
	x = max(bounds.Min.X, min(x, bounds.Max.X-textWidth))
	y = max(bounds.Min.Y+ascent, min(y, bounds.Max.Y))
	return fixed.P(x, y)
}

// drawLine uses a two-pixel stroke so scaled diagonals remain continuous.
func drawLine(canvas *image.RGBA, from, to point, colour color.Color) {
	pixel := color.RGBAModel.Convert(colour).(color.RGBA)
	x0, y0 := int(math.Round(from.x)), int(math.Round(from.y))
	x1, y1 := int(math.Round(to.x)), int(math.Round(to.y))
	dx, dy := x1-x0, y1-y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	stepX, stepY := -1, -1
	if x0 < x1 {
		stepX = 1
	}
	if y0 < y1 {
		stepY = 1
	}

	err := dx - dy
	for {
		canvas.SetRGBA(x0, y0, pixel)
		if dx >= dy {
			canvas.SetRGBA(x0, y0+1, pixel)
		} else {
			canvas.SetRGBA(x0+1, y0, pixel)
		}
		if x0 == x1 && y0 == y1 {
			return
		}
		twice := 2 * err
		if twice > -dy {
			err -= dy
			x0 += stepX
		}
		if twice < dx {
			err += dx
			y0 += stepY
		}
	}
}

func pixelPoint(position point) point {
	position = screenPoint(position)
	return point{
		x: position.x*pixelsPerCellX + pixelsPerCellX/2,
		y: position.y*pixelsPerCellY + pixelsPerCellY/2,
	}
}

// screenPoint keeps graph coordinates independent from viewport padding.
func screenPoint(position point) point {
	return point{x: position.x + graphInsetX, y: position.y + graphInsetY}
}

func drawCircle(canvas *image.RGBA, center point, radius float64, colour color.Color, opacity float64) {
	pixel := color.NRGBAModel.Convert(colour).(color.NRGBA)
	minimumX := int(math.Floor(center.x - radius - 1))
	maximumX := int(math.Ceil(center.x + radius + 1))
	minimumY := int(math.Floor(center.y - radius - 1))
	maximumY := int(math.Ceil(center.y + radius + 1))
	for y := minimumY; y <= maximumY; y++ {
		for x := minimumX; x <= maximumX; x++ {
			distance := math.Hypot(float64(x)+0.5-center.x, float64(y)+0.5-center.y)
			coverage := clamp(radius+0.5-distance, 0, 1) * opacity
			if coverage > 0 {
				blend(canvas, x, y, pixel, coverage)
			}
		}
	}
}

func blend(canvas *image.RGBA, x, y int, source color.NRGBA, opacity float64) {
	if !image.Pt(x, y).In(canvas.Bounds()) {
		return
	}
	sourceAlpha := uint32(float64(source.A) * clamp(opacity, 0, 1))
	inverse := 255 - sourceAlpha
	offset := canvas.PixOffset(x, y)
	canvas.Pix[offset] = uint8((uint32(source.R)*sourceAlpha + uint32(canvas.Pix[offset])*inverse) / 255)
	canvas.Pix[offset+1] = uint8((uint32(source.G)*sourceAlpha + uint32(canvas.Pix[offset+1])*inverse) / 255)
	canvas.Pix[offset+2] = uint8((uint32(source.B)*sourceAlpha + uint32(canvas.Pix[offset+2])*inverse) / 255)
	canvas.Pix[offset+3] = uint8(sourceAlpha + uint32(canvas.Pix[offset+3])*inverse/255)
}
