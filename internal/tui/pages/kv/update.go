package kv

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	redismetrics "github.com/briheet/sen/internal/tui/pages/kv/metrics"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
)

var toggleMetrics = key.NewBinding(key.WithKeys("M", "shift+m"))

// Init starts commands required by the key-value page.
func (Model) Init() tea.Cmd { return nil }

// Update handles messages for the key-value page.
func (m Model) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if _, ok := msg.(tea.BackgroundColorMsg); ok {
		m.metrics, _ = m.metrics.Update(msg)
		m.revision++
	}
	if obscured, ok := msg.(pages.ObscuredMsg); ok {
		m.obscured = obscured.Obscured
		obscured.Obscured = m.obscured || m.showMetrics
		var command tea.Cmd
		m.graph, command = m.graph.Update(obscured)
		m.revision++
		return m, command
	}
	if tick, ok := msg.(pages.TelemetryTickMsg); ok && m.Engine != nil {
		revision := m.Engine.Revision()
		if revision == m.telemetry {
			return m, nil
		}
		snapshot := m.Engine.Snapshot()
		addCommandEdges(m.Engine.Graph.Static, &snapshot)
		if m.dump != nil {
			_, _ = fmt.Fprintf(m.dump, "kv[%s] telemetry revision=%d active_nodes=%d active_edges=%d\n",
				m.Service.Name, revision, len(snapshot.NodeActivity), len(snapshot.NodeEdges))
		}
		m.telemetry = revision
		m.metrics.ApplySnapshot(snapshot.Metrics, tick.At)
		m.revision++
		var command tea.Cmd
		m.graph, command = m.graph.Update(graph.TelemetryMsg{
			Nodes:     snapshot.NodeActivity,
			Files:     snapshot.FileActivity,
			NodeEdges: snapshot.NodeEdges,
			FileEdges: snapshot.FileEdges,
		})
		return m, command
	}
	if press, ok := msg.(tea.KeyPressMsg); ok && m.Engine != nil && key.Matches(press, toggleMetrics) {
		m.showMetrics = !m.showMetrics
		m.revision++
		var command tea.Cmd
		m.graph, command = m.graph.Update(pages.ObscuredMsg{Obscured: m.showMetrics || m.obscured})
		return m, command
	}
	if m.showMetrics {
		if press, ok := msg.(tea.KeyPressMsg); ok && press.Code == tea.KeyEscape {
			m.showMetrics = false
			m.revision++
			var command tea.Cmd
			m.graph, command = m.graph.Update(pages.ObscuredMsg{Obscured: m.obscured})
			return m, command
		}
		switch msg.(type) {
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg, tea.MouseWheelMsg:
			return m, nil
		case tea.KeyPressMsg:
			m.metrics, _ = m.metrics.Update(msg)
			m.revision++
			return m, nil
		}
	}
	if viewport, ok := msg.(pages.ViewportMsg); ok {
		m.width = max(0, viewport.Width)
		m.height = max(0, viewport.Height)
		panelWidth, panelHeight := redismetrics.Size(m.width, m.height)
		m.metrics, _ = m.metrics.Update(tea.WindowSizeMsg{Width: panelWidth, Height: panelHeight})
		var command tea.Cmd
		m.graph, command = m.graph.Update(viewport)
		m.revision++
		return m, command
	}

	switch msg.(type) {
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg, tea.MouseWheelMsg:
		var command tea.Cmd
		m.graph, command = m.graph.Update(msg)
		return m, command
	}

	var command tea.Cmd
	m.graph, command = m.graph.Update(msg)
	return m, command
}

// addCommandEdges maps profile activity onto Redis's synthetic root-to-command
// edges. Server graphs continue to use only edges observed in runtime traces.
func addCommandEdges(static *model.StaticGraph, snapshot *model.RuntimeSnapshot) {
	if static == nil || snapshot == nil {
		return
	}
	root := static.Nodes[static.Root]
	if root == nil {
		return
	}
	for _, command := range root.Out {
		if activity := snapshot.NodeActivity[command]; activity > 0 {
			snapshot.NodeEdges[model.NodeEdge{From: root.ID, To: command}] = activity
		}
	}
}
