package db

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

// View renders the selected database graph and optional metrics dashboard.
func (m *Model) View() tea.View {
	revision := m.Revision()
	if m.viewValid && m.viewRevision == revision {
		return m.view
	}
	bodyHeight := max(0, m.height-1)
	content := m.graphs[m.pager.Page].View()
	if m.Engine == nil {
		// Connection strings commonly contain passwords, so never render Address.
		message := m.Service.Name + " (" + string(m.Service.Provider) + ")\n\nTelemetry unavailable."
		content = lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, message)
	}
	if m.showMetrics {
		if panel := m.metrics.View(); panel != "" {
			content = pages.Overlay(&m.canvas, m.width, bodyHeight, content, panel)
		}
	}
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, m.indicatorView()))
	view.MouseMode = tea.MouseModeCellMotion
	m.view, m.viewRevision, m.viewValid = view, revision, true
	return m.view
}

func (m *Model) indicatorView() string {
	if m.indicator == "" || m.indicatorPage != m.pager.Page || m.indicatorWidth != m.width {
		m.indicator = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.pager.View())
		m.indicatorPage, m.indicatorWidth = m.pager.Page, m.width
	}
	return m.indicator
}
