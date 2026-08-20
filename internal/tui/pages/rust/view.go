package rust

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// View renders the selected Rust graph and optional metrics dashboard.
func (m Model) View() tea.View {
	content := m.graphs[m.pager.Page].View()
	if m.showMetrics {
		panel := m.metrics.View()
		if panel != "" {
			bodyHeight := max(0, m.height-1)
			canvas := lipgloss.NewCanvas(m.width, bodyHeight)
			base := lipgloss.NewLayer(content)
			overlay := lipgloss.NewLayer(panel).
				X(max(0, (m.width-lipgloss.Width(panel))/2)).
				Y(max(0, (bodyHeight-lipgloss.Height(panel))/2))
			canvas.Compose(lipgloss.NewCompositor(base, overlay))
			content = canvas.Render()
		}
	}
	indicator := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.pager.View())
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, indicator))
	view.MouseMode = tea.MouseModeCellMotion
	return view
}
