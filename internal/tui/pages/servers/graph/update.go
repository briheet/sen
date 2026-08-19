package graph

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

const (
	maximumReturnFrames = 120
	baseEdgeLength      = 20.0
	returnEdgeStrength  = 0.7
	collisionStrength   = 16.0
	nodeHalfWidth       = 1.0
	nodeHalfHeight      = 1.5
	nodeWidth           = 2
)

type frameMsg struct {
	owner      string
	generation uint64
}

type uploadMsg struct {
	owner         string
	width, height int
}

func nextFrame(owner string, generation uint64) tea.Cmd {
	return tea.Tick(time.Second/framesPerSecond, func(time.Time) tea.Msg {
		return frameMsg{owner: owner, generation: generation}
	})
}

func nextUpload(owner string, width, height int) tea.Cmd {
	return tea.Tick(time.Second/framesPerSecond, func(time.Time) tea.Msg {
		return uploadMsg{owner: owner, width: width, height: height}
	})
}

// Update handles graph resizing, dragging, and layout frames.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pages.ViewportMsg:
		wasVisible := m.visible
		m.originX, m.originY = msg.X, msg.Y
		m.visible = msg.Visible
		m.resize(msg.Width, msg.Height)
		m.renderPending = false
		m.trace("viewport origin=%d,%d size=%dx%d cell=%dx%d visible=%t",
			m.originX, m.originY, m.width, m.height, m.cellWidth, m.cellHeight, m.visible)
		if !m.visible {
			m.animating = false
			m.generation++
			m.renderSequence++
			m.renderer.cancel(m.renderSequence)
			m.frontImageID = 0
			if wasVisible && m.graphics {
				m.trace("deactivated images=%d,%d", m.imageIDs[0], m.imageIDs[1])
				return m, deleteImagesCommand(m.imageIDs, m.quiet())
			}
			return m, nil
		}
		if !m.graphics || len(m.nodes) == 0 {
			return m, nil
		}
		m.trace("activated")
		// Let Bubble Tea paint native labels before placing graph pixels.
		return m, nextUpload(m.owner, m.width, m.height)
	case uploadMsg:
		if msg.owner != m.owner || msg.width != m.width || msg.height != m.height || !m.visible {
			return m, nil
		}
		return m, m.upload()
	case renderReadyMsg:
		frame := msg.frame
		if frame.owner != m.owner || frame.sequence != m.renderSequence || !m.visible {
			releaseRenderNodes(frame.buffer)
			return m, nil
		}
		m.refreshRenderedLabels(false)
		return m, tea.Raw(frame)
	case tea.RawMsg:
		frame, ok := msg.Msg.(frameOutput)
		if !ok || frame.owner != m.owner {
			return m, nil
		}
		if frame.sequence == m.renderSequence && m.visible {
			m.frontImageID = frame.imageID
			m.renderPending = false
			releaseRenderNodes(frame.buffer)
			m.trace("frame committed image_id=%d sequence=%d", frame.imageID, frame.sequence)
			if m.animating {
				return m, nextFrame(m.owner, m.generation)
			}
			return m, nil
		}
		releaseRenderNodes(frame.buffer)
		return m, deleteImagesCommand(m.imageIDs, m.quiet())
	case renderFailedMsg:
		if msg.owner == m.owner {
			m.renderErr = msg.err
			m.renderPending = false
			m.animating = false
			m.revision++
		}
		return m, nil
	case tea.MouseClickMsg:
		if !m.graphics || !m.visible {
			return m, nil
		}
		if msg.Button == tea.MouseLeft && m.beginDrag(msg.X, msg.Y) {
			m.trace("drag started node=%q x=%d y=%d", m.nodes[m.dragging].label, msg.X, msg.Y)
			m.animating = true
			m.layoutFrames = 0
			m.generation++
			m.refreshLabels(false)
			return m, m.upload()
		}
		if msg.Button == tea.MouseLeft {
			m.trace("drag missed x=%d y=%d", msg.X, msg.Y)
		}
	case tea.MouseMotionMsg:
		if m.graphics && m.visible && m.dragging >= 0 {
			m.moveDragged(msg.X, msg.Y)
			// The frame clock renders the latest position and coalesces mouse events.
			return m, nil
		}
	case tea.MouseReleaseMsg:
		if m.graphics && m.visible && msg.Button == tea.MouseLeft && m.dragging >= 0 {
			m.moveDragged(msg.X, msg.Y)
			m.trace("drag stopped node=%q x=%d y=%d", m.nodes[m.dragging].label, msg.X, msg.Y)
			m.dragging = -1
			m.layoutFrames = 0
			m.refreshLabels(false)
			return m, nil
		}
	case frameMsg:
		if msg.owner != m.owner || msg.generation != m.generation || !m.animating {
			return m, nil
		}
		m.stepAnchored()
		if m.dragging < 0 {
			m.layoutFrames++
		}

		if m.dragging < 0 && (m.settled() || m.layoutFrames >= maximumReturnFrames) {
			m.restoreAnchors()
			m.animating = false
			return m, m.upload()
		}
		// The committed frame schedules the next tick, providing render backpressure.
		return m, m.upload()
	}
	return m, nil
}

func (m *Model) resize(width, height int) {
	m.width = max(0, width)
	m.height = max(0, height)
	cellWidth, cellHeight := terminalCellSize()
	if cellWidth != m.cellWidth || cellHeight != m.cellHeight {
		m.cellWidth = cellWidth
		m.cellHeight = cellHeight
		m.nodeRadius = max(3, float64(cellWidth)*0.65)
	}
	m.layoutFrames = 0
	if len(m.nodes) == 0 || m.width == 0 || m.height == 0 {
		m.animating = false
		m.refreshLabels(true)
		return
	}

	maximumDepth := 0
	for _, node := range m.nodes {
		maximumDepth = max(maximumDepth, node.depth)
	}
	availableWidth := float64(max(0, m.width-nodeWidth-int(graphInsetX)))
	availableHeight := float64(max(0, m.height-2-int(graphInsetY)))
	for index := range m.nodes {
		node := &m.nodes[index]
		if index == m.root {
			node.position = point{}
		} else {
			node.position = point{
				x: availableWidth * float64(node.depth) / float64(max(1, maximumDepth)),
				y: availableHeight * float64(node.row+1) / float64(node.rowCount+1),
			}
		}
		node.position = m.clamp(node.position)
		node.velocity = point{}
	}
	m.pinRoot()
	for index := range m.edges {
		edge := &m.edges[index]
		edge.rest = max(baseEdgeLength, distance(m.nodes[edge.from].position, m.nodes[edge.to].position))
	}
	m.dragging = -1
	m.animating = false
	m.captureAnchors()
	m.trace("layout distributed width=%.0f height=%.0f", availableWidth, availableHeight)
	m.refreshLabels(true)
}

func (m *Model) stepAnchored() {
	forces := m.forces()
	const deltaTime = 1.0 / framesPerSecond
	for index := range m.nodes {
		if index == m.root || index == m.dragging {
			continue
		}
		node := &m.nodes[index]
		node.velocity.x = clamp(node.velocity.x+forces[index].x*deltaTime, -maxVelocity, maxVelocity)
		node.velocity.y = clamp(node.velocity.y+forces[index].y*deltaTime, -maxVelocity, maxVelocity)
		node.position.x, node.velocity.x = m.spring.Update(node.position.x, node.velocity.x, node.anchor.x)
		node.position.y, node.velocity.y = m.spring.Update(node.position.y, node.velocity.y, node.anchor.y)
		node.position = m.clamp(node.position)
	}
	m.pinRoot()
}

func (m *Model) forces() []point {
	if cap(m.forceBuffer) < len(m.nodes) {
		m.forceBuffer = make([]point, len(m.nodes))
	} else {
		m.forceBuffer = m.forceBuffer[:len(m.nodes)]
		clear(m.forceBuffer)
	}
	forces := m.forceBuffer
	for _, edge := range m.edges {
		if edge.from == edge.to {
			continue
		}
		first := m.nodes[edge.from].position
		second := m.nodes[edge.to].position
		delta := point{x: second.x - first.x, y: second.y - first.y}
		length := distance(first, second)
		if length == 0 {
			delta = point{x: 1, y: 0}
			length = 1
		}
		force := returnEdgeStrength * (length - edge.rest)
		m.addPairForce(forces, edge.from, edge.to, point{x: force * delta.x / length, y: force * delta.y / length})
	}
	for first := range m.nodes {
		for second := first + 1; second < len(m.nodes); second++ {
			m.addCollisionForces(forces, first, second)
		}
	}
	return forces
}

func (m *Model) addPairForce(forces []point, first, second int, force point) {
	forces[first].x += force.x
	forces[first].y += force.y
	forces[second].x -= force.x
	forces[second].y -= force.y
}

func (m *Model) addCollisionForces(forces []point, first, second int) {
	firstCenter := m.nodes[first].position
	secondCenter := m.nodes[second].position
	delta := point{x: secondCenter.x - firstCenter.x, y: secondCenter.y - firstCenter.y}
	if delta == (point{}) {
		delta = point{x: 1, y: 0.5}
	}
	overlapX := nodeHalfWidth*(m.nodes[first].scale+m.nodes[second].scale) - math.Abs(delta.x)
	overlapY := nodeHalfHeight*(m.nodes[first].scale+m.nodes[second].scale) - math.Abs(delta.y)
	if overlapX <= 0 || overlapY <= 0 {
		return
	}
	force := point{}
	if overlapX < overlapY {
		force.x = -math.Copysign(collisionStrength*overlapX, delta.x)
	} else {
		force.y = -math.Copysign(collisionStrength*overlapY, delta.y)
	}
	m.addPairForce(forces, first, second, force)
}

func (m *Model) captureAnchors() {
	for index := range m.nodes {
		node := &m.nodes[index]
		node.anchor = node.position
		node.velocity = point{}
	}
	m.pinRoot()
}

func (m *Model) restoreAnchors() {
	for index := range m.nodes {
		m.nodes[index].position = m.nodes[index].anchor
		m.nodes[index].velocity = point{}
	}
	m.pinRoot()
}

func (m *Model) pinRoot() {
	if m.root < 0 || m.root >= len(m.nodes) {
		return
	}
	root := &m.nodes[m.root]
	root.position = point{x: 0, y: 0}
	root.velocity = point{}
	root.anchor = root.position
}

func (m *Model) beginDrag(x, y int) bool {
	for index := len(m.nodes) - 1; index >= 0; index-- {
		if index == m.root {
			continue
		}
		node := &m.nodes[index]
		position := screenPoint(node.position)
		nodeX := int(math.Round(position.x))
		nodeY := int(math.Round(position.y))
		radiusX := max(1, int(math.Ceil(m.nodeRadius*node.scale/float64(m.cellWidth))))
		radiusY := max(1, int(math.Ceil(m.nodeRadius*node.scale/float64(m.cellHeight))))
		if y >= nodeY-radiusY && y <= nodeY+radiusY && x >= nodeX-radiusX && x <= nodeX+radiusX {
			m.dragging = index
			m.dragOffset = point{x: float64(x) - position.x, y: float64(y) - position.y}
			node.velocity = point{}
			return true
		}
	}
	return false
}

func (m *Model) moveDragged(x, y int) {
	node := &m.nodes[m.dragging]
	next := m.clamp(point{
		x: float64(x) - graphInsetX - m.dragOffset.x,
		y: float64(y) - graphInsetY - m.dragOffset.y,
	})
	delta := point{x: next.x - node.position.x, y: next.y - node.position.y}
	if delta != (point{}) {
		node.velocity = point{
			x: clamp(delta.x*framesPerSecond, -maxVelocity, maxVelocity),
			y: clamp(delta.y*framesPerSecond, -maxVelocity, maxVelocity),
		}
		node.position = next
	}
}

func (m *Model) settled() bool {
	for index, node := range m.nodes {
		if index == m.root {
			continue
		}
		if distance(node.position, node.anchor) > settleDistance ||
			math.Abs(node.velocity.x) > settleDistance || math.Abs(node.velocity.y) > settleDistance {
			return false
		}
	}
	return true
}

func (m *Model) clamp(position point) point {
	minimumX := 0.0
	maximumX := float64(max(int(minimumX), m.width-nodeWidth-int(graphInsetX)))
	minimumY := 0.0
	maximumY := float64(max(int(minimumY), m.height-2-int(graphInsetY)))
	return point{
		x: clamp(position.x, minimumX, maximumX),
		y: clamp(position.y, minimumY, maximumY),
	}
}

func clamp(value, minimum, maximum float64) float64 {
	return min(max(value, minimum), maximum)
}
