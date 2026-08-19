package graph

import (
	"image"
	"image/color"
	"math"

	"github.com/briheet/sen/internal/tui/styles"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	fallbackCellWidth  = 6
	fallbackCellHeight = 16
	fallbackNodeRadius = 4
	fallbackLabelSize  = 7.0
	graphInsetX        = 1.0
	graphInsetY        = 1.0
	alphaLevels        = 16
)

type graphPalette struct {
	colors color.Palette
	text   [alphaLevels + 1]uint8
	idle   [alphaLevels + 1]uint8
	active [alphaLevels + 1]uint8
	hot    [alphaLevels + 1]uint8
}

// textCanvas maps glyph alpha directly to the text palette ramp.
type textCanvas struct {
	canvas  *image.Paletted
	indexes *[alphaLevels + 1]uint8
}

func (c textCanvas) ColorModel() color.Model { return c.canvas.ColorModel() }
func (c textCanvas) Bounds() image.Rectangle { return c.canvas.Bounds() }
func (c textCanvas) At(x, y int) color.Color { return c.canvas.At(x, y) }

func (c textCanvas) RGBA64At(x, y int) color.RGBA64 {
	red, green, blue, alpha := c.canvas.At(x, y).RGBA()
	return color.RGBA64{R: uint16(red), G: uint16(green), B: uint16(blue), A: uint16(alpha)}
}

func (c textCanvas) Set(x, y int, colour color.Color) {
	_, _, _, alpha := colour.RGBA()
	c.setAlpha(x, y, alpha)
}

func (c textCanvas) SetRGBA64(x, y int, colour color.RGBA64) {
	c.setAlpha(x, y, uint32(colour.A))
}

func (c textCanvas) setAlpha(x, y int, alpha uint32) {
	level := int((alpha*alphaLevels + 0x7fff) / 0xffff)
	if level > 0 {
		c.canvas.SetColorIndex(x, y, c.indexes[min(level, alphaLevels)])
	}
}

func newGraphPalette(theme styles.Theme) graphPalette {
	palette := graphPalette{colors: color.Palette{color.NRGBA{}}}
	add := func(colour color.Color, indexes *[alphaLevels + 1]uint8) {
		pixel := color.NRGBAModel.Convert(colour).(color.NRGBA)
		for level := 1; level <= alphaLevels; level++ {
			indexes[level] = uint8(len(palette.colors))
			pixel.A = uint8(level * 255 / alphaLevels)
			palette.colors = append(palette.colors, pixel)
		}
	}
	add(theme.Text, &palette.text)
	add(theme.NodeIdle, &palette.idle)
	add(theme.NodeActive, &palette.active)
	add(theme.NodeHot, &palette.hot)
	return palette
}

var labelFont = func() *opentype.Font {
	parsed, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	return parsed
}()

func newLabelFace(size float64) font.Face {
	face, err := opentype.NewFace(labelFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
	return face
}

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
	if m.renderErr != nil {
		return centeredMessage(m.width, m.height, "Unable to render graph.")
	}
	if m.placeholder == "" {
		return centeredMessage(m.width, m.height, "Graph viewport is too large.")
	}
	return m.placeholder
}

func (r *renderer) renderImage(request renderRequest) *image.Paletted {
	bounds := image.Rect(0, 0, request.width*request.cellWidth, request.height*request.cellHeight)
	if r.canvas == nil || r.canvas.Bounds() != bounds {
		r.canvas = image.NewPaletted(bounds, r.palette.colors)
	} else {
		clear(r.canvas.Pix)
	}
	canvas := r.canvas
	for _, edge := range r.edges {
		from, to := edgePoints(request.nodes[edge.from], request.nodes[edge.to], request.cellWidth, request.cellHeight, request.nodeRadius)
		edge.draw(canvas, from, to, r.palette.idle[alphaLevels])
	}
	labels := textCanvas{canvas: canvas, indexes: &r.palette.text}
	drawer := font.Drawer{
		Dst:  labels,
		Src:  image.NewUniform(r.palette.colors[r.palette.text[alphaLevels]]),
		Face: request.face,
	}
	for index, node := range request.nodes {
		position := pixelPoint(node.position, request.cellWidth, request.cellHeight)
		indexes := &r.palette.active
		if index == request.dragging {
			indexes = &r.palette.hot
		}
		drawCircle(canvas, position, request.nodeRadius+2, indexes, 0.15)
		drawCircle(canvas, position, request.nodeRadius, indexes, 1)
		drawer.Dot = labelPoint(canvas.Bounds(), position, node.label, request.face, request.nodeRadius)
		drawer.DrawString(node.label)
	}
	return canvas
}

func (m edgeModel) draw(canvas *image.Paletted, from, to point, colour uint8) {
	drawLine(canvas, from, to, colour)
}

// edgePoints connects circle boundaries; labels are a separate render layer.
func edgePoints(from, to node, cellWidth, cellHeight int, radius float64) (point, point) {
	start := pixelPoint(from.position, cellWidth, cellHeight)
	end := pixelPoint(to.position, cellWidth, cellHeight)
	delta := point{x: end.x - start.x, y: end.y - start.y}
	length := distance(start, end)
	if length == 0 {
		return start, end
	}
	offset := min(max(1, radius-1), length/2)
	start.x += delta.x / length * offset
	start.y += delta.y / length * offset
	end.x -= delta.x / length * offset
	end.y -= delta.y / length * offset
	return start, end
}

func labelPoint(bounds image.Rectangle, position point, label string, face font.Face, radius float64) fixed.Point26_6 {
	textWidth := font.MeasureString(face, label).Ceil()
	gap := max(2, radius/2)
	x := int(math.Round(position.x)) - textWidth/2
	y := int(math.Round(position.y - radius - gap))
	ascent := face.Metrics().Ascent.Ceil()
	if y < ascent {
		x = int(math.Round(position.x + radius + gap))
		y = int(math.Round(position.y)) + ascent/2
	}
	x = max(bounds.Min.X, min(x, bounds.Max.X-textWidth))
	y = max(bounds.Min.Y+ascent, min(y, bounds.Max.Y))
	return fixed.P(x, y)
}

// drawLine uses a two-pixel stroke so scaled diagonals remain continuous.
func drawLine(canvas *image.Paletted, from, to point, colour uint8) {
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
		canvas.SetColorIndex(x0, y0, colour)
		if dx >= dy {
			canvas.SetColorIndex(x0, y0+1, colour)
		} else {
			canvas.SetColorIndex(x0+1, y0, colour)
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

func pixelPoint(position point, cellWidth, cellHeight int) point {
	position = screenPoint(position)
	return point{
		x: position.x*float64(cellWidth) + float64(cellWidth)/2,
		y: position.y*float64(cellHeight) + float64(cellHeight)/2,
	}
}

// screenPoint keeps graph coordinates independent from viewport padding.
func screenPoint(position point) point {
	return point{x: position.x + graphInsetX, y: position.y + graphInsetY}
}

func drawCircle(canvas *image.Paletted, center point, radius float64, indexes *[alphaLevels + 1]uint8, opacity float64) {
	minimumX := int(math.Floor(center.x - radius - 1))
	maximumX := int(math.Ceil(center.x + radius + 1))
	minimumY := int(math.Floor(center.y - radius - 1))
	maximumY := int(math.Ceil(center.y + radius + 1))
	for y := minimumY; y <= maximumY; y++ {
		for x := minimumX; x <= maximumX; x++ {
			distance := math.Hypot(float64(x)+0.5-center.x, float64(y)+0.5-center.y)
			coverage := clamp(radius+0.5-distance, 0, 1) * opacity
			level := int(math.Round(coverage * alphaLevels))
			if level > 0 {
				canvas.SetColorIndex(x, y, indexes[min(level, alphaLevels)])
			}
		}
	}
}
