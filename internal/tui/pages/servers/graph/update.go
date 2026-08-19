package graph

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
)

const (
	minimumLayoutFrames = 30
	maximumLayoutFrames = 90
	maximumReturnFrames = 120
	baseEdgeLength      = 20.0
	layoutEdgeStrength  = 1.8
	returnEdgeStrength  = 0.7
	seedStrength        = 0.35
	repulsionDistance   = 12.0
	repulsionStrength   = 20.0
	collisionStrength   = 16.0
	velocityDamping     = 0.88
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
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		m.trace("resized width=%d height=%d cell=%dx%d",
			m.width, m.height, m.cellWidth, m.cellHeight)
		if !m.graphics || len(m.nodes) == 0 {
			return m, nil
		}
		// Let Bubble Tea enter the alternate screen before transmitting pixels.
		return m, nextUpload(m.owner, m.width, m.height)
	case uploadMsg:
		if msg.owner != m.owner || msg.width != m.width || msg.height != m.height {
			return m, nil
		}
		return m, m.upload()
	case tea.RawMsg:
		// The renderer has handed the encoded frame back to Bubble Tea.
		m.trace("upload accepted")
		return m, nil
	case renderFailedMsg:
		if msg.owner == m.owner {
			m.renderErr = msg.err
		}
		return m, nil
	case tea.MouseClickMsg:
		if !m.graphics {
			return m, nil
		}
		if msg.Button == tea.MouseLeft && m.beginDrag(msg.X, msg.Y) {
			m.trace("drag started node=%q x=%d y=%d", m.nodes[m.dragging].label, msg.X, msg.Y)
			m.animating = true
			m.layoutFrames = 0
			m.generation++
			return m, tea.Batch(m.upload(), nextFrame(m.owner, m.generation))
		}
		if msg.Button == tea.MouseLeft {
			m.trace("drag missed x=%d y=%d", msg.X, msg.Y)
		}
	case tea.MouseMotionMsg:
		if m.graphics && m.dragging >= 0 {
			m.moveDragged(msg.X, msg.Y)
			// The frame clock renders the latest position and coalesces mouse events.
			return m, nil
		}
	case tea.MouseReleaseMsg:
		if m.graphics && msg.Button == tea.MouseLeft && m.dragging >= 0 {
			m.moveDragged(msg.X, msg.Y)
			m.trace("drag stopped node=%q x=%d y=%d", m.nodes[m.dragging].label, msg.X, msg.Y)
			m.dragging = -1
			m.layoutFrames = 0
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

		upload := m.upload()
		if m.dragging < 0 && (m.settled() || m.layoutFrames >= maximumReturnFrames) {
			m.restoreAnchors()
			m.animating = false
			return m, upload
		}
		return m, tea.Batch(upload, nextFrame(m.owner, m.generation))
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
		m.labelFace = newLabelFace(max(8, float64(cellHeight)*0.55))
	}
	m.placeholder = ""
	if m.graphics && m.width > 0 && m.height > 0 {
		m.placeholder, _ = placeholders(m.imageID, m.width, m.height)
	}
	m.layoutFrames = 0
	if len(m.nodes) == 0 || m.width == 0 || m.height == 0 {
		m.animating = false
		return
	}

	maximumDepth := 0
	for _, node := range m.nodes {
		maximumDepth = max(maximumDepth, node.depth)
	}
	edgeLength := min(baseEdgeLength, max(4, float64(m.width-nodeWidth-int(graphInsetX))/float64(max(1, maximumDepth))))
	rowGap := min(8, max(4, edgeLength/3))
	for index := range m.nodes {
		node := &m.nodes[index]
		if index == m.root {
			node.seed = point{x: 0, y: 0}
		} else {
			node.seed = point{
				x: float64(node.depth) * edgeLength,
				y: 2 + float64(node.row)*rowGap,
			}
		}
		node.position = m.clamp(node.seed)
		node.anchor = point{}
		node.velocity = point{}
	}
	m.pinRoot()
	for index := range m.edges {
		m.edges[index].rest = edgeLength
	}
	m.dragging = -1
	if len(m.nodes) > 1 {
		m.settleInitialLayout()
	} else {
		m.captureAnchors()
	}
}

// settleInitialLayout calculates stable anchors without encoding intermediate frames.
func (m *Model) settleInitialLayout() {
	frames := 0
	for frames = 1; frames <= maximumLayoutFrames; frames++ {
		motion := m.stepLayout()
		if frames >= minimumLayoutFrames && motion < settleDistance {
			break
		}
	}
	m.captureAnchors()
	m.animating = false
	m.layoutFrames = 0
	m.trace("layout settled frames=%d", min(frames, maximumLayoutFrames))
}

func (m *Model) stepLayout() float64 {
	forces := m.forces(true)
	maximumMotion := 0.0
	const deltaTime = 1.0 / framesPerSecond
	for index := range m.nodes {
		if index == m.root {
			continue
		}
		node := &m.nodes[index]
		node.velocity.x = clamp((node.velocity.x+forces[index].x*deltaTime)*velocityDamping, -maxVelocity, maxVelocity)
		node.velocity.y = clamp((node.velocity.y+forces[index].y*deltaTime)*velocityDamping, -maxVelocity, maxVelocity)
		movement := point{x: node.velocity.x * deltaTime, y: node.velocity.y * deltaTime}
		node.position = m.clamp(point{x: node.position.x + movement.x, y: node.position.y + movement.y})
		maximumMotion = max(maximumMotion, distance(point{}, movement))
	}
	m.pinRoot()
	return maximumMotion
}

func (m *Model) stepAnchored() {
	forces := m.forces(false)
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

func (m *Model) forces(layout bool) []point {
	if cap(m.forceBuffer) < len(m.nodes) {
		m.forceBuffer = make([]point, len(m.nodes))
	} else {
		m.forceBuffer = m.forceBuffer[:len(m.nodes)]
		clear(m.forceBuffer)
	}
	forces := m.forceBuffer
	edgeStrength := returnEdgeStrength
	if layout {
		edgeStrength = layoutEdgeStrength
	}
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
		force := edgeStrength * (length - edge.rest)
		m.addPairForce(forces, edge.from, edge.to, point{x: force * delta.x / length, y: force * delta.y / length})
	}
	if layout {
		for index, node := range m.nodes {
			forces[index].x += seedStrength * (node.seed.x - node.position.x)
			forces[index].y += seedStrength * (node.seed.y - node.position.y)
		}
	}
	for first := range m.nodes {
		for second := first + 1; second < len(m.nodes); second++ {
			m.addCollisionForces(forces, first, second, layout)
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

func (m *Model) addCollisionForces(forces []point, first, second int, repel bool) {
	firstCenter := m.nodes[first].position
	secondCenter := m.nodes[second].position
	delta := point{x: secondCenter.x - firstCenter.x, y: secondCenter.y - firstCenter.y}
	length := distance(firstCenter, secondCenter)
	if length == 0 {
		delta = point{x: 1, y: 0.5}
		length = distance(point{}, delta)
	}
	if repel && length < repulsionDistance {
		strength := repulsionStrength * (1 - length/repulsionDistance)
		m.addPairForce(forces, first, second, point{x: -strength * delta.x / length, y: -strength * delta.y / length})
	}

	overlapX := 2*nodeHalfWidth - math.Abs(delta.x)
	overlapY := 2*nodeHalfHeight - math.Abs(delta.y)
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
		if y >= nodeY-1 && y <= nodeY+1 && x >= nodeX-1 && x <= nodeX+1 {
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
