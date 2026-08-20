package tigerbeetle

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// View renders the selected graph and optional metrics dashboard.
func (m Model) View() tea.View {
	bodyHeight := max(0, m.height-1)
	content := m.graphs[m.pager.Page].View()
	if m.Engine == nil {
		message := m.Service.Name + " (tigerbeetle)\n\nTelemetry unavailable."
		content = lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, message)
	}
	if m.showMetrics {
		if panel := m.metrics.View(); panel != "" {
			canvas := lipgloss.NewCanvas(m.width, bodyHeight)
			canvas.Compose(lipgloss.NewCompositor(
				lipgloss.NewLayer(content),
				lipgloss.NewLayer(panel).X(max(0, (m.width-lipgloss.Width(panel))/2)).Y(max(0, (bodyHeight-lipgloss.Height(panel))/2)),
			))
			content = canvas.Render()
		}
	}
	indicator := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.pager.View())
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, indicator))
	view.MouseMode = tea.MouseModeCellMotion
	return view
}
