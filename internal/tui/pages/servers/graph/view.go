package graph

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
	"golang.org/x/image/vector"
)

const (
	fallbackCellWidth  = 6
	fallbackCellHeight = 16
	fallbackNodeRadius = 4
	graphInsetX        = 3.0
	graphInsetY        = 1.0
	alphaLevels        = 63
	edgeWidth          = 1.5
	labelHysteresis    = 0.75
)

type graphPalette struct {
	colors color.Palette
	idle   [alphaLevels + 1]uint8
	active [alphaLevels + 1]uint8
	hot    [alphaLevels + 1]uint8
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
	add(theme.NodeIdle, &palette.idle)
	add(theme.NodeActive, &palette.active)
	add(theme.NodeHot, &palette.hot)
	return palette
}

// View returns native terminal labels; graph pixels are placed behind them.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if len(m.nodes) == 0 {
		if m.kind == FileGraph {
			return centeredMessage(m.width, m.height, "No project files found.")
		}
		return centeredMessage(m.width, m.height, "No project functions found.")
	}
	if !m.graphics {
		return centeredMessage(m.width, m.height, "Pixel graph requires Ghostty or Kitty.")
	}
	if m.renderErr != nil {
		return centeredMessage(m.width, m.height, "Unable to render graph.")
	}
	return m.labels
}

// refreshLabels rebuilds terminal text only after a label crosses a cell.
func (m *Model) refreshLabels(force bool) {
	m.refreshLabelPositions(force, false)
}

func (m *Model) refreshRenderedLabels(force bool) {
	m.refreshLabelPositions(force, true)
}

func (m *Model) refreshLabelPositions(force, rendered bool) {
	redraw := force
	for index := range m.nodes {
		node := &m.nodes[index]
		position := node.position
		if rendered {
			position = node.rendered
		}
		position = screenPoint(position)
		cellX := stableLabelCell(position.x, node.labelCellX, force)
		cellY := stableLabelCell(position.y, node.labelCellY, force)
		x, y := m.labelPosition(node.label, cellX, cellY)
		if x != node.labelX || y != node.labelY || cellX != node.labelCellX || cellY != node.labelCellY {
			node.labelX, node.labelY = x, y
			node.labelCellX, node.labelCellY = cellX, cellY
			redraw = true
		}
	}
	if m.labelsDragging != m.dragging {
		m.labelsDragging = m.dragging
		redraw = true
	}
	if !redraw {
		return
	}

	cells := make([]rune, m.width*m.height)
	hot := make([]bool, len(cells))
	for index := range cells {
		cells[index] = ' '
	}
	for index, node := range m.nodes {
		for offset, character := range []rune(node.label) {
			x := node.labelX + offset
			if x < 0 || x >= m.width || node.labelY < 0 || node.labelY >= m.height {
				continue
			}
			position := node.labelY*m.width + x
			cells[position] = character
			hot[position] = index == m.dragging
		}
	}

	normalStyle := lipgloss.NewStyle().Foreground(styles.Zakura.Text)
	hotStyle := lipgloss.NewStyle().Foreground(styles.Zakura.NodeHot)
	var output strings.Builder
	output.Grow(len(cells) + m.height)
	for y := range m.height {
		row := cells[y*m.width : (y+1)*m.width]
		rowHot := hot[y*m.width : (y+1)*m.width]
		for start := 0; start < len(row); {
			if row[start] == ' ' {
				output.WriteRune(' ')
				start++
				continue
			}
			end := start + 1
			for end < len(row) && row[end] != ' ' && rowHot[end] == rowHot[start] {
				end++
			}
			style := normalStyle
			if rowHot[start] {
				style = hotStyle
			}
			output.WriteString(style.Render(string(row[start:end])))
			start = end
		}
		if y < m.height-1 {
			output.WriteByte('\n')
		}
	}
	m.labels = output.String()
	m.revision++
}

func (m Model) labelPosition(label string, cellX, cellY int) (int, int) {
	labelWidth := len([]rune(label))
	x := cellX - labelWidth/2
	y := cellY - 1
	if y < 0 {
		x = cellX + 1
		y = cellY
	}
	x = max(0, min(x, max(0, m.width-labelWidth)))
	y = max(0, min(y, max(0, m.height-1)))
	return x, y
}

// stableLabelCell prevents native text from flickering across a cell boundary.
func stableLabelCell(position float64, current int, force bool) int {
	if force || math.Abs(position-float64(current)) > labelHysteresis {
		return int(math.Round(position))
	}
	return current
}

func (r *renderer) renderImage(request renderRequest) *image.Paletted {
	bounds := image.Rect(0, 0, request.width*request.cellWidth, request.height*request.cellHeight)
	if r.canvas == nil || r.canvas.Bounds() != bounds {
		r.canvas = image.NewPaletted(bounds, r.palette.colors)
		r.mask = image.NewAlpha(bounds)
	} else {
		clear(r.canvas.Pix)
	}

	// Glows sit below links so translucent pixels cannot erase connections.
	r.drawNodes(request, false, 2, 0.15)
	r.drawNodes(request, true, 2, 0.2)
	r.drawEdges(request, false)
	r.drawEdges(request, true)
	r.drawNodes(request, false, 0, 1)
	r.drawNodes(request, true, 0, 1)
	return r.canvas
}

func (r *renderer) drawEdges(request renderRequest, hot bool) {
	drawn := false
	for _, edge := range r.edges {
		connected := edge.from == request.dragging || edge.to == request.dragging
		if connected != hot {
			continue
		}
		if !drawn {
			r.resetMask()
			drawn = true
		}
		from := pixelPoint(request.nodes[edge.from].position, request.cellWidth, request.cellHeight)
		to := pixelPoint(request.nodes[edge.to].position, request.cellWidth, request.cellHeight)
		addCapsule(&r.rasterizer, from, to, edgeWidth)
	}
	if !drawn {
		return
	}
	indexes := &r.palette.idle
	if hot {
		indexes = &r.palette.hot
	}
	r.paintMask(indexes, 1)
}

func (r *renderer) drawNodes(request renderRequest, hotOnly bool, radiusOffset, opacity float64) {
	drawn := false
	for index, node := range request.nodes {
		if (index == request.dragging) != hotOnly {
			continue
		}
		if !drawn {
			r.resetMask()
			drawn = true
		}
		radius := request.nodeRadius*node.scale + radiusOffset
		addCircle(&r.rasterizer, pixelPoint(node.position, request.cellWidth, request.cellHeight), radius)
	}
	if !drawn {
		return
	}
	indexes := &r.palette.active
	if hotOnly {
		indexes = &r.palette.hot
	}
	r.paintMask(indexes, opacity)
}

func (r *renderer) resetMask() {
	clear(r.mask.Pix)
	r.rasterizer.Reset(r.mask.Bounds().Dx(), r.mask.Bounds().Dy())
	r.rasterizer.DrawOp = draw.Src
}

func (r *renderer) paintMask(indexes *[alphaLevels + 1]uint8, opacity float64) {
	r.rasterizer.Draw(r.mask, r.mask.Bounds(), image.Opaque, image.Point{})
	for index, alpha := range r.mask.Pix {
		level := int(math.Round(float64(alpha) * opacity * alphaLevels / 255))
		if level > 0 {
			r.canvas.Pix[index] = indexes[min(level, alphaLevels)]
		}
	}
}

func addCapsule(rasterizer *vector.Rasterizer, from, to point, width float64) {
	delta := point{x: to.x - from.x, y: to.y - from.y}
	length := math.Hypot(delta.x, delta.y)
	radius := width / 2
	if length == 0 {
		addCircle(rasterizer, from, radius)
		return
	}
	normal := point{x: -delta.y / length * radius, y: delta.x / length * radius}
	rasterizer.MoveTo(float32(from.x+normal.x), float32(from.y+normal.y))
	rasterizer.LineTo(float32(to.x+normal.x), float32(to.y+normal.y))
	rasterizer.LineTo(float32(to.x-normal.x), float32(to.y-normal.y))
	rasterizer.LineTo(float32(from.x-normal.x), float32(from.y-normal.y))
	rasterizer.ClosePath()
	addCircle(rasterizer, from, radius)
	addCircle(rasterizer, to, radius)
}

func addCircle(rasterizer *vector.Rasterizer, center point, radius float64) {
	const kappa = 0.5522847498307936
	x, y, control := center.x, center.y, radius*kappa
	rasterizer.MoveTo(float32(x+radius), float32(y))
	rasterizer.CubeTo(float32(x+radius), float32(y+control), float32(x+control), float32(y+radius), float32(x), float32(y+radius))
	rasterizer.CubeTo(float32(x-control), float32(y+radius), float32(x-radius), float32(y+control), float32(x-radius), float32(y))
	rasterizer.CubeTo(float32(x-radius), float32(y-control), float32(x-control), float32(y-radius), float32(x), float32(y-radius))
	rasterizer.CubeTo(float32(x+control), float32(y-radius), float32(x+radius), float32(y-control), float32(x+radius), float32(y))
	rasterizer.ClosePath()
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
