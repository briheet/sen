package graph

import (
	"cmp"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
)

const (
	fallbackCellWidth  = 6
	fallbackCellHeight = 16
	fallbackNodeRadius = 4
	alphaLevels        = 63
	edgeWidth          = 1.25
	labelHysteresis    = 0.75
	minimumLabelZoom   = 0.6
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

func (m *Model) refreshLabels(force bool) {
	m.refreshLabelPositions(force, false)
}

func (m *Model) refreshRenderedLabels(force bool) {
	m.refreshLabelPositions(force, true)
}

// refreshLabelPositions admits important labels without allowing overlaps.
func (m *Model) refreshLabelPositions(force, rendered bool) {
	if m.width == 0 || m.height == 0 {
		return
	}
	camera := m.camera
	selected, hovered, dragging := m.selected, m.hovered, m.dragging
	if rendered {
		camera = m.renderedCamera
		selected, hovered, dragging = m.renderedSelect, m.renderedHover, m.renderedDrag
	}
	m.labelScratch.focused = reuse(m.labelScratch.focused, len(m.nodes))
	focused := m.labelScratch.focused
	if selected >= 0 && selected < len(focused) {
		focused[selected] = true
		for _, edge := range m.edges {
			if edge.from == selected {
				focused[edge.to] = true
			}
			if edge.to == selected {
				focused[edge.from] = true
			}
		}
	}
	m.labelScratch.order = reuse(m.labelScratch.order, len(m.nodes))
	order := m.labelScratch.order
	for index := range order {
		order[index] = index
	}
	slices.SortFunc(order, func(left, right int) int {
		leftPriority := m.labelPriority(left, selected, hovered, dragging, focused)
		rightPriority := m.labelPriority(right, selected, hovered, dragging, focused)
		if leftPriority != rightPriority {
			return rightPriority - leftPriority
		}
		if m.nodes[left].degree != m.nodes[right].degree {
			return m.nodes[right].degree - m.nodes[left].degree
		}
		return cmp.Compare(m.nodes[left].id, m.nodes[right].id)
	})

	redraw := force || m.labelsDragging != dragging || m.labelsSelected != selected || m.labelsHovered != hovered
	m.labelScratch.occupied = reuse(m.labelScratch.occupied, m.width*m.height)
	occupied := m.labelScratch.occupied
	for _, index := range order {
		node := &m.nodes[index]
		position := node.position
		if rendered {
			position = node.rendered
		}
		screen := camera.worldToScreen(position)
		cellX := stableLabelCell(screen.x/float64(m.cellWidth)-0.5, node.labelCellX, force)
		cellY := stableLabelCell(screen.y/float64(m.cellHeight)-0.5, node.labelCellY, force)
		priority := m.labelPriority(index, selected, hovered, dragging, focused)
		visible := priority > 0 || camera.zoom >= minimumLabelZoom
		x, y := 0, 0
		if visible {
			x, y, visible = m.placeLabel(node.labelWidth, cellX, cellY, occupied)
		}
		if x != node.labelX || y != node.labelY || cellX != node.labelCellX || cellY != node.labelCellY || visible != node.labelVisible {
			node.labelX, node.labelY = x, y
			node.labelCellX, node.labelCellY = cellX, cellY
			node.labelVisible = visible
			redraw = true
		}
	}
	if !redraw {
		return
	}
	m.labelsDragging, m.labelsSelected, m.labelsHovered = dragging, selected, hovered
	m.buildLabelView(selected, hovered, dragging, focused)
}

func (m *Model) labelPriority(index, selected, hovered, dragging int, focused []bool) int {
	switch {
	case index == selected || index == dragging:
		return 5
	case index == hovered:
		return 4
	case index == m.root:
		return 3
	case selected >= 0 && focused[index]:
		return 2
	default:
		return 0
	}
}

func (m *Model) placeLabel(width, cellX, cellY int, occupied []bool) (int, int, bool) {
	candidates := [3][2]int{
		{cellX - width/2, cellY - 1},
		{cellX - width/2, cellY + 1},
		{cellX + 1, cellY},
	}
	for _, candidate := range candidates {
		x, y := candidate[0], candidate[1]
		if x < 0 || y < 0 || y >= m.height || x+width > m.width {
			continue
		}
		free := true
		for offset := range width {
			if occupied[y*m.width+x+offset] {
				free = false
				break
			}
		}
		if !free {
			continue
		}
		for offset := range width {
			occupied[y*m.width+x+offset] = true
		}
		return x, y, true
	}
	return 0, 0, false
}

func (m *Model) buildLabelView(selected, hovered, dragging int, focused []bool) {
	m.labelScratch.cells = reuse(m.labelScratch.cells, m.width*m.height)
	m.labelScratch.classes = reuse(m.labelScratch.classes, len(m.labelScratch.cells))
	cells, classes := m.labelScratch.cells, m.labelScratch.classes
	for index := range cells {
		cells[index] = ' '
	}
	for index := range m.nodes {
		node := &m.nodes[index]
		if !node.labelVisible {
			continue
		}
		class := uint8(1)
		if index == selected || index == hovered || index == dragging {
			class = 2
		} else if selected >= 0 && !focused[index] {
			class = 0
		}
		offset := 0
		for _, character := range node.label {
			position := node.labelY*m.width + node.labelX + offset
			cells[position], classes[position] = character, class
			offset++
		}
	}

	mutedStyle := lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted)
	normalStyle := lipgloss.NewStyle().Foreground(styles.Zakura.Text)
	hotStyle := lipgloss.NewStyle().Foreground(styles.Zakura.NodeHot)
	if m.obscured {
		mutedStyle = lipgloss.NewStyle().Foreground(styles.Zakura.Border).Faint(true)
		normalStyle = mutedStyle
		hotStyle = mutedStyle
	}
	var output strings.Builder
	output.Grow(len(cells) + m.height)
	for y := range m.height {
		row := cells[y*m.width : (y+1)*m.width]
		rowClasses := classes[y*m.width : (y+1)*m.width]
		for start := 0; start < len(row); {
			if row[start] == ' ' {
				output.WriteRune(' ')
				start++
				continue
			}
			end := start + 1
			for end < len(row) && row[end] != ' ' && rowClasses[end] == rowClasses[start] {
				end++
			}
			style := mutedStyle
			if rowClasses[start] == 1 {
				style = normalStyle
			} else if rowClasses[start] == 2 {
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

func reuse[T any](values []T, size int) []T {
	if cap(values) < size {
		return make([]T, size)
	}
	values = values[:size]
	clear(values)
	return values
}

// stableLabelCell prevents native text from flickering across cell boundaries.
func stableLabelCell(position float64, current int, force bool) int {
	if force || math.Abs(position-float64(current)) > labelHysteresis {
		return int(math.Round(position))
	}
	return current
}

func (r *renderer) renderImage(request *renderRequest, canvas *image.Paletted) {
	clear(canvas.Pix)
	r.classifyNodes(request)

	r.drawNodeClass(canvas, request, 2, 3, 0.16)
	r.drawEdgeClass(canvas, request, 0, 0.22)
	r.drawEdgeClass(canvas, request, 1, 0.48)
	r.drawEdgeClass(canvas, request, 2, 0.95)
	r.drawNodeClass(canvas, request, 0, 0, 0.34)
	r.drawNodeClass(canvas, request, 1, 0, 0.9)
	r.drawNodeClass(canvas, request, 2, 0, 1)
}

func (r *renderer) drawEdgeClass(canvas *image.Paletted, request *renderRequest, class int, opacity float64) {
	indexes := &r.palette.idle
	if class == 1 {
		indexes = &r.palette.active
	} else if class == 2 {
		indexes = &r.palette.hot
	}
	for _, edge := range request.edges {
		if edgeClass(request, edge) != class {
			continue
		}
		from := request.camera.worldToScreen(request.nodes[edge.from].position)
		to := request.camera.worldToScreen(request.nodes[edge.to].position)
		fromRadius := renderNodeRadius(request, &request.nodes[edge.from])
		toRadius := renderNodeRadius(request, &request.nodes[edge.to])
		from, to = clipEdge(from, to, fromRadius, toRadius)
		drawCapsule(canvas, indexes, from, to, edgeWidth, opacity)
	}
}

func edgeClass(request *renderRequest, edge edgeModel) int {
	if request.selected >= 0 {
		if edge.from == request.selected || edge.to == request.selected {
			return 2
		}
		return 0
	}
	hot := request.dragging
	if hot < 0 {
		hot = request.hovered
	}
	if hot >= 0 && (edge.from == hot || edge.to == hot) {
		return 2
	}
	if edge.active {
		return 2
	}
	return 1
}

func (r *renderer) drawNodeClass(canvas *image.Paletted, request *renderRequest, class int, radiusOffset, opacity float64) {
	indexes := &r.palette.idle
	if class == 1 {
		indexes = &r.palette.active
	} else if class == 2 {
		indexes = &r.palette.hot
	}
	for index := range request.nodes {
		if int(r.nodeClass[index]) != class {
			continue
		}
		node := &request.nodes[index]
		drawCircle(canvas, indexes, request.camera.worldToScreen(node.position), renderNodeRadius(request, node)+radiusOffset, opacity)
	}
}

func (r *renderer) classifyNodes(request *renderRequest) {
	if cap(r.nodeClass) < len(request.nodes) {
		r.nodeClass = make([]uint8, len(request.nodes))
	} else {
		r.nodeClass = r.nodeClass[:len(request.nodes)]
	}
	if request.selected < 0 {
		for index := range r.nodeClass {
			r.nodeClass[index] = 1
		}
	} else {
		clear(r.nodeClass)
		for _, edge := range request.edges {
			if edge.from == request.selected {
				r.nodeClass[edge.to] = 1
			}
			if edge.to == request.selected {
				r.nodeClass[edge.from] = 1
			}
		}
	}
	for _, index := range [...]int{request.dragging, request.selected, request.hovered} {
		if index >= 0 && index < len(r.nodeClass) {
			r.nodeClass[index] = 2
		}
	}
	for index := range request.nodes {
		if request.nodes[index].active {
			r.nodeClass[index] = 2
		}
	}
}

func renderNodeRadius(request *renderRequest, node *node) float64 {
	zoomScale := clamp(math.Sqrt(request.camera.zoom), 0.75, 1.5)
	return request.nodeRadius * node.scale * zoomScale
}

func clipEdge(from, to point, fromRadius, toRadius float64) (point, point) {
	delta := point{x: to.x - from.x, y: to.y - from.y}
	length := math.Hypot(delta.x, delta.y)
	if length <= fromRadius+toRadius || length == 0 {
		return from, to
	}
	direction := point{x: delta.x / length, y: delta.y / length}
	return point{x: from.x + direction.x*fromRadius, y: from.y + direction.y*fromRadius},
		point{x: to.x - direction.x*toRadius, y: to.y - direction.y*toRadius}
}

// drawCapsule follows the dominant axis, touching only pixels near the edge.
func drawCapsule(canvas *image.Paletted, indexes *[alphaLevels + 1]uint8, from, to point, width, opacity float64) {
	delta := point{x: to.x - from.x, y: to.y - from.y}
	lengthSquared := delta.x*delta.x + delta.y*delta.y
	radius := width / 2
	if delta == (point{}) {
		drawCircle(canvas, indexes, from, radius, opacity)
		return
	}
	drawCircle(canvas, indexes, from, radius, opacity)
	drawCircle(canvas, indexes, to, radius, opacity)
	padding := int(math.Ceil(radius + 0.5))
	if math.Abs(delta.x) >= math.Abs(delta.y) {
		left, right := int(math.Floor(min(from.x, to.x))), int(math.Ceil(max(from.x, to.x)))
		for x := left; x <= right; x++ {
			t := clamp((float64(x)+0.5-from.x)/delta.x, 0, 1)
			y := from.y + delta.y*t
			for pixelY := int(math.Floor(y)) - padding; pixelY <= int(math.Floor(y))+padding; pixelY++ {
				paintDistance(canvas, indexes, x, pixelY, pointSegmentDistance(point{x: float64(x) + 0.5, y: float64(pixelY) + 0.5}, from, delta, lengthSquared), radius, opacity)
			}
		}
		return
	}
	top, bottom := int(math.Floor(min(from.y, to.y))), int(math.Ceil(max(from.y, to.y)))
	for y := top; y <= bottom; y++ {
		t := clamp((float64(y)+0.5-from.y)/delta.y, 0, 1)
		x := from.x + delta.x*t
		for pixelX := int(math.Floor(x)) - padding; pixelX <= int(math.Floor(x))+padding; pixelX++ {
			paintDistance(canvas, indexes, pixelX, y, pointSegmentDistance(point{x: float64(pixelX) + 0.5, y: float64(y) + 0.5}, from, delta, lengthSquared), radius, opacity)
		}
	}
}

func drawCircle(canvas *image.Paletted, indexes *[alphaLevels + 1]uint8, center point, radius, opacity float64) {
	padding := radius + 0.5
	left, right := int(math.Floor(center.x-padding)), int(math.Ceil(center.x+padding))
	top, bottom := int(math.Floor(center.y-padding)), int(math.Ceil(center.y+padding))
	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			distance := math.Hypot(float64(x)+0.5-center.x, float64(y)+0.5-center.y)
			paintDistance(canvas, indexes, x, y, distance, radius, opacity)
		}
	}
}

func paintDistance(canvas *image.Paletted, indexes *[alphaLevels + 1]uint8, x, y int, distance, radius, opacity float64) {
	if !image.Pt(x, y).In(canvas.Bounds()) {
		return
	}
	coverage := clamp(radius+0.5-distance, 0, 1)
	level := int(math.Round(coverage * opacity * alphaLevels))
	if level == 0 {
		return
	}
	position := canvas.PixOffset(x, y)
	current := canvas.Pix[position]
	if current >= indexes[1] && current <= indexes[alphaLevels] && current >= indexes[level] {
		return
	}
	canvas.Pix[position] = indexes[level]
}

func pointSegmentDistance(position, from, delta point, lengthSquared float64) float64 {
	if lengthSquared == 0 {
		return distance(position, from)
	}
	t := clamp(((position.x-from.x)*delta.x+(position.y-from.y)*delta.y)/lengthSquared, 0, 1)
	closest := point{x: from.x + delta.x*t, y: from.y + delta.y*t}
	return distance(position, closest)
}
