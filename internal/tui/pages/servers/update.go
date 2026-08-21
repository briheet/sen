// This file deals with server model implementation
package servers

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/pages/servers/metrics"
)

var toggleMetrics = key.NewBinding(key.WithKeys("M", "shift+m"))
var previousGraph = key.NewBinding(key.WithKeys("H", "shift+h"))
var nextGraph = key.NewBinding(key.WithKeys("L", "shift+l"))

// switchGraphMsg selects a graph after the previous Kitty image is removed.
type switchGraphMsg struct {
	service string
	page    int
}

// Init starts commands required by the server page.
func (m Model) Init() tea.Cmd {
	commands := make([]tea.Cmd, 0, len(m.graphs)+1)
	commands = append(commands, m.metrics.Init())
	for index := range m.graphs {
		if command := m.graphs[index].Init(); command != nil {
			commands = append(commands, command)
		}
	}
	return tea.Batch(commands...)
}

// Update handles messages for the server page.
func (m Model) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if _, ok := msg.(tea.BackgroundColorMsg); ok {
		m.metrics, _ = m.metrics.Update(msg)
		m.revision++
	}
	if obscured, ok := msg.(pages.ObscuredMsg); ok {
		m.obscured = obscured.Obscured
		commands := make([]tea.Cmd, 0, len(m.graphs))
		for index := range m.graphs {
			obscured.Obscured = m.obscured || (m.showMetrics && index == m.pager.Page)
			var command tea.Cmd
			m.graphs[index], command = m.graphs[index].Update(obscured)
			if command != nil {
				commands = append(commands, command)
			}
		}
		m.revision++
		return m, tea.Batch(commands...)
	}
	if tick, ok := msg.(pages.TelemetryTickMsg); ok {
		revision := m.Engine.Revision()
		if revision == m.telemetry {
			return m, nil
		}
		snapshot := m.Engine.Snapshot()
		m.telemetry = revision
		m.metrics.ApplySnapshot(snapshot.Metrics, tick.At)
		m.revision++
		commands := make([]tea.Cmd, 0, len(m.graphs))
		for index := range m.graphs {
			var command tea.Cmd
			m.graphs[index], command = m.graphs[index].Update(graph.TelemetryMsg{
				Nodes:     snapshot.NodeActivity,
				Files:     snapshot.FileActivity,
				NodeEdges: snapshot.NodeEdges,
				FileEdges: snapshot.FileEdges,
			})
			if command != nil {
				commands = append(commands, command)
			}
		}
		return m, tea.Batch(commands...)
	}
	if press, ok := msg.(tea.KeyPressMsg); ok && key.Matches(press, toggleMetrics) {
		m.showMetrics = !m.showMetrics
		m.revision++
		var command tea.Cmd
		m.graphs[m.pager.Page], command = m.graphs[m.pager.Page].Update(pages.ObscuredMsg{Obscured: m.showMetrics || m.obscured})
		return m, command
	}
	if m.showMetrics {
		if press, ok := msg.(tea.KeyPressMsg); ok && press.Code == tea.KeyEscape {
			m.showMetrics = false
			m.revision++
			var command tea.Cmd
			m.graphs[m.pager.Page], command = m.graphs[m.pager.Page].Update(pages.ObscuredMsg{Obscured: m.obscured})
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
	if change, ok := msg.(switchGraphMsg); ok {
		if change.service != m.Name() || change.page != m.pending {
			return m, nil
		}
		m.pending = -1
		m.pager.Page = change.page
		m.revision++
		viewport := m.viewport
		viewport.Width = m.width
		viewport.Height = max(0, m.height-1)
		viewport.Visible = m.viewport.Visible
		var command tea.Cmd
		m.graphs[change.page], command = m.graphs[change.page].Update(viewport)
		return m, command
	}
	if press, ok := msg.(tea.KeyPressMsg); ok && m.pending < 0 {
		switch {
		case key.Matches(press, previousGraph):
			if page, command, changed := m.switchPage(m.pager.Page - 1); changed {
				return page, command
			}
		case key.Matches(press, nextGraph):
			if page, command, changed := m.switchPage(m.pager.Page + 1); changed {
				return page, command
			}
		}
	}

	if viewport, ok := msg.(pages.ViewportMsg); ok {
		m.viewport = viewport
		m.width = max(0, viewport.Width)
		m.height = max(0, viewport.Height)
		panelWidth, panelHeight := metrics.Size(m.width, max(0, m.height-1))
		m.metrics, _ = m.metrics.Update(tea.WindowSizeMsg{Width: panelWidth, Height: panelHeight})
		commands := make([]tea.Cmd, 0, len(m.graphs))
		for index := range m.graphs {
			graphViewport := viewport
			graphViewport.Width = m.width
			graphViewport.Height = max(0, m.height-1)
			graphViewport.Visible = viewport.Visible && index == m.pager.Page && m.pending < 0
			var command tea.Cmd
			m.graphs[index], command = m.graphs[index].Update(graphViewport)
			if command != nil {
				commands = append(commands, command)
			}
		}
		return m, tea.Batch(commands...)
	}

	if click, ok := msg.(tea.MouseClickMsg); ok && click.Y == m.height-1 {
		indicator := m.pager.View()
		start := max(0, (m.width-lipgloss.Width(indicator))/2)
		if offset := click.X - start; click.Button == tea.MouseLeft && offset >= 0 && offset < lipgloss.Width(indicator) {
			page := min(m.pager.TotalPages-1, offset/lipgloss.Width(m.pager.ActiveDot))
			if page != m.pager.Page && m.pending < 0 {
				previous := m.pager.Page
				m.pending = page
				viewport := m.viewport
				viewport.Width = m.width
				viewport.Height = max(0, m.height-1)

				hidden := viewport
				hidden.Visible = false
				var hideCommand tea.Cmd
				m.graphs[previous], hideCommand = m.graphs[previous].Update(hidden)

				showCommand := func() tea.Msg {
					return switchGraphMsg{service: m.Name(), page: page}
				}
				// Select the next page only after deletion, so Bubble Tea paints its
				// labels before the new Kitty image is placed behind them.
				return m, tea.Sequence(hideCommand, showCommand)
			}
		}
		return m, nil
	}
	m.pager, _ = m.pager.Update(msg)
	switch msg.(type) {
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg, tea.MouseWheelMsg:
		var command tea.Cmd
		m.graphs[m.pager.Page], command = m.graphs[m.pager.Page].Update(msg)
		return m, command
	}

	// Internal render messages are owner-tagged, so every graph can safely consume them.
	commands := make([]tea.Cmd, 0, len(m.graphs))
	for index := range m.graphs {
		var command tea.Cmd
		m.graphs[index], command = m.graphs[index].Update(msg)
		if command != nil {
			commands = append(commands, command)
		}
	}
	return m, tea.Batch(commands...)
}

func (m Model) switchPage(page int) (pages.Page, tea.Cmd, bool) {
	if page < 0 || page >= m.pager.TotalPages || page == m.pager.Page || m.pending >= 0 {
		return m, nil, false
	}
	previous := m.pager.Page
	m.pending = page
	viewport := m.viewport
	viewport.Width = m.width
	viewport.Height = max(0, m.height-1)

	hidden := viewport
	hidden.Visible = false
	var hideCommand tea.Cmd
	m.graphs[previous], hideCommand = m.graphs[previous].Update(hidden)

	showCommand := func() tea.Msg {
		return switchGraphMsg{service: m.Name(), page: page}
	}
	return m, tea.Sequence(hideCommand, showCommand), true
}
