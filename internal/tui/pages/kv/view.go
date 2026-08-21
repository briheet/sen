package kv

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

// View renders the selected key-value service.
func (m *Model) View() tea.View {
	revision := m.Revision()
	if m.viewValid && m.viewRevision == revision {
		return m.view
	}
	bodyHeight := max(0, m.height-1)
	content := m.graph.View()
	if m.Engine == nil {
		message := m.Service.Name + " (" + string(m.Service.Provider) + ")\n" + m.Service.Address + "\n\nTelemetry unavailable."
		content = lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, message)
	}
	if m.showMetrics {
		panel := m.metrics.View()
		if panel != "" {
			content = pages.Overlay(&m.canvas, m.width, bodyHeight, content, panel)
		}
	}
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, m.indicatorView()))
	view.MouseMode = tea.MouseModeCellMotion
	m.view, m.viewRevision, m.viewValid = view, revision, true
	return m.view
}

func (m *Model) indicatorView() string {
	if m.indicator == "" || m.indicatorWidth != m.width {
		m.indicator = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.pager.View())
		m.indicatorWidth = m.width
	}
	return m.indicator
}
