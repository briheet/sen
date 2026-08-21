package graph

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
)

type frameMsg struct {
	owner      string
	generation uint64
}

type uploadMsg struct {
	owner         string
	width, height int
}

// TelemetryMsg applies activity from the latest completed trace window.
type TelemetryMsg struct {
	Nodes     map[model.NodeID]int64
	Files     map[model.FileID]int64
	NodeEdges map[model.NodeEdge]int64
	FileEdges map[model.FileEdge]int64
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

// Update handles graph rendering, navigation, focus, and force-layout frames.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TelemetryMsg:
		changed := false
		for index := range m.nodes {
			active := msg.Nodes[model.NodeID(m.nodes[index].id)] > 0
			if m.kind == FileGraph {
				active = msg.Files[model.FileID(m.nodes[index].id)] > 0
			}
			if m.nodes[index].active != active {
				m.nodes[index].active = active
				changed = true
			}
		}
		for index := range m.edges {
			edge := &m.edges[index]
			from, to := m.nodes[edge.from].id, m.nodes[edge.to].id
			active := msg.NodeEdges[model.NodeEdge{From: model.NodeID(from), To: model.NodeID(to)}] > 0
			if m.kind == FileGraph {
				active = msg.FileEdges[model.FileEdge{From: model.FileID(from), To: model.FileID(to)}] > 0
			}
			if edge.active != active {
				edge.active = active
				changed = true
			}
		}
		if !changed {
			return m, nil
		}
		m.dirty = true
		return m, m.upload()
	case pages.ObscuredMsg:
		if m.obscured == msg.Obscured {
			return m, nil
		}
		m.obscured = msg.Obscured
		m.refreshLabels(true)
		m.dirty = true
		return m, m.upload()
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
			if m.renderer != nil {
				m.renderer.cancel(m.renderSequence)
			}
			m.frontImageID = 0
			if wasVisible && m.graphics {
				return m, deleteImagesCommand(m.imageIDs, m.quiet())
			}
			return m, nil
		}
		if !m.graphics || len(m.nodes) == 0 {
			return m, nil
		}
		m.dirty = true
		// Paint native labels before placing graph pixels beneath them.
		return m, nextUpload(m.owner, m.width, m.height)
	case uploadMsg:
		if msg.owner != m.owner {
			return m, nil
		}
		m.uploadScheduled = false
		if msg.width != m.width || msg.height != m.height || !m.visible {
			return m, nil
		}
		// Pointer traffic can keep the renderer dirty continuously, delaying the
		// separate animation timer. Advance linked nodes here so every displayed
		// drag frame reflects the latest force response.
		if m.dragging >= 0 && m.animating {
			m.advanceSimulation(draggingStepsPerFrame)
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
			if m.dirty {
				return m, m.scheduleUpload()
			}
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
	case tea.KeyPressMsg:
		if !m.visible {
			return m, nil
		}
		switch msg.String() {
		case "esc":
			if m.selected >= 0 {
				m.selected = -1
				m.dirty = true
				return m, m.upload()
			}
		case "0":
			m.camera.manual = false
			m.camera.fit(m.nodes)
			m.dirty = true
			return m, m.upload()
		}
	case tea.MouseWheelMsg:
		if !m.graphics || !m.visible {
			return m, nil
		}
		factor := zoomStep
		if msg.Button == tea.MouseWheelDown {
			factor = 1 / zoomStep
		} else if msg.Button != tea.MouseWheelUp {
			return m, nil
		}
		if !m.camera.zoomAt(m.mousePixel(msg.X, msg.Y), factor) {
			return m, nil
		}
		return m, m.scheduleUpload()
	case tea.MouseClickMsg:
		if !m.graphics || !m.visible || msg.Button != tea.MouseLeft {
			return m, nil
		}
		pointer := m.mousePixel(msg.X, msg.Y)
		m.pressPoint, m.lastPointer = pointer, pointer
		m.pointerMoved = false
		m.pressed = m.hitNode(pointer)
		m.panning = m.pressed < 0
		return m, nil
	case tea.MouseMotionMsg:
		if !m.graphics || !m.visible {
			return m, nil
		}
		pointer := m.mousePixel(msg.X, msg.Y)
		if m.pressed >= 0 && msg.Button == tea.MouseLeft {
			if m.dragging < 0 && distance(pointer, m.pressPoint) >= float64(min(m.cellWidth, m.cellHeight)) {
				m.startDrag()
			}
			if m.dragging >= 0 {
				m.moveDragged(pointer)
				m.pointerMoved = true
				return m, m.scheduleUpload()
			}
		}
		if m.panning && msg.Button == tea.MouseLeft {
			delta := point{x: pointer.x - m.lastPointer.x, y: pointer.y - m.lastPointer.y}
			m.camera.pan(delta)
			m.lastPointer = pointer
			m.pointerMoved = m.pointerMoved || delta != (point{})
			return m, m.scheduleUpload()
		}
		hovered := m.hitNode(pointer)
		if hovered != m.hovered {
			m.hovered = hovered
			return m, m.scheduleUpload()
		}
	case tea.MouseReleaseMsg:
		if !m.graphics || !m.visible || msg.Button != tea.MouseLeft {
			return m, nil
		}
		pointer := m.mousePixel(msg.X, msg.Y)
		switch {
		case m.dragging >= 0:
			m.moveDragged(pointer)
			dragged := m.dragging
			m.nodes[dragged].fixed = false
			m.dragging, m.pressed = -1, -1
			m.simulation.cool()
			m.animating = true
			return m, m.scheduleUpload()
		case m.pressed >= 0:
			selected := m.pressed
			m.pressed = -1
			m.selected = selected
			return m, m.scheduleUpload()
		case m.panning:
			m.panning = false
			if !m.pointerMoved {
				m.selected = -1
			}
			// Replace the low-resolution interaction frame with full quality.
			return m, m.scheduleUpload()
		}
	case frameMsg:
		if msg.owner != m.owner || msg.generation != m.generation || !m.animating {
			return m, nil
		}
		steps := settlingStepsPerFrame
		if m.dragging >= 0 {
			steps = draggingStepsPerFrame
		}
		// Advance several physics ticks per displayed frame so linked nodes keep
		// pace without increasing PNG uploads or terminal traffic.
		settled := m.advanceSimulation(steps)
		if settled && m.dragging < 0 {
			m.animating = false
		}
		if m.visualHash() == m.renderHash {
			if m.animating {
				return m, nextFrame(m.owner, m.generation)
			}
			return m, nil
		}
		m.dirty = true
		return m, m.upload()
	}
	return m, nil
}

func (m *Model) advanceSimulation(steps int) bool {
	settled := false
	for range steps {
		settled = m.simulation.step(m.nodes, m.edges)
		if settled {
			break
		}
	}
	return settled
}

// scheduleUpload coalesces high-frequency pointer input behind one frame
// deadline. The latest state wins, so interaction remains responsive without
// encoding a complete PNG for every mouse event.
func (m *Model) scheduleUpload() tea.Cmd {
	m.dirty = true
	if !m.graphics || !m.visible || m.renderPending || m.uploadScheduled ||
		len(m.nodes) == 0 || m.width == 0 || m.height == 0 {
		return nil
	}
	m.uploadScheduled = true
	return nextUpload(m.owner, m.width, m.height)
}

func (m *Model) resize(width, height int) {
	oldWidth, oldHeight := m.width, m.height
	m.width, m.height = max(0, width), max(0, height)
	cellWidth, cellHeight := terminalCellSize()
	if cellWidth != m.cellWidth || cellHeight != m.cellHeight {
		m.cellWidth, m.cellHeight = cellWidth, cellHeight
		m.nodeRadius = max(3, float64(cellWidth)*0.55)
	}
	m.camera.resize(m.width*m.cellWidth, m.height*m.cellHeight)
	if !m.camera.manual || oldWidth == 0 || oldHeight == 0 {
		m.camera.fit(m.nodes)
	}
	if m.renderedCamera.zoom == 0 {
		m.renderedCamera = m.camera
	}
	m.refreshLabels(true)
}

func (m *Model) startDrag() {
	if m.pressed < 0 {
		return
	}
	m.dragging = m.pressed
	node := &m.nodes[m.dragging]
	world := m.camera.screenToWorld(m.pressPoint)
	m.dragOffset = point{x: world.x - node.position.x, y: world.y - node.position.y}
	node.fixed = true
	node.velocity = point{}
	m.simulation.reheat()
	m.animating = true
	m.generation++
}

func (m *Model) moveDragged(pointer point) {
	if m.dragging < 0 {
		return
	}
	world := m.camera.screenToWorld(pointer)
	node := &m.nodes[m.dragging]
	node.position = point{x: world.x - m.dragOffset.x, y: world.y - m.dragOffset.y}
	node.velocity = point{}
}

func (m Model) mousePixel(x, y int) point {
	return point{
		x: float64(x*m.cellWidth) + float64(m.cellWidth)/2,
		y: float64(y*m.cellHeight) + float64(m.cellHeight)/2,
	}
}

func (m Model) hitNode(pointer point) int {
	for index := len(m.nodes) - 1; index >= 0; index-- {
		center := m.camera.worldToScreen(m.nodes[index].position)
		radius := m.nodePixelRadius(m.nodes[index]) + float64(min(m.cellWidth, m.cellHeight))/2
		if distance(center, pointer) <= radius {
			return index
		}
	}
	return -1
}

func (m Model) nodePixelRadius(node node) float64 {
	return m.nodeRadius * node.scale * nodeZoomScale(m.camera.zoom)
}

func clamp(value, minimum, maximum float64) float64 {
	return min(max(value, minimum), maximum)
}
