package kv

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// View renders the selected key-value service.
func (m Model) View() tea.View {
	content := m.graph.View()
	if m.Engine == nil {
		message := m.Service.Name + " (" + string(m.Service.Provider) + ")\n" + m.Service.Address + "\n\nTelemetry unavailable."
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, message)
	}
	if m.showMetrics {
		panel := m.metrics.View()
		if panel != "" {
			canvas := lipgloss.NewCanvas(m.width, m.height)
			canvas.Compose(lipgloss.NewCompositor(
				lipgloss.NewLayer(content),
				lipgloss.NewLayer(panel).
					X(max(0, (m.width-lipgloss.Width(panel))/2)).
					Y(max(0, (m.height-lipgloss.Height(panel))/2)),
			))
			content = canvas.Render()
		}
	}
	view := tea.NewView(content)
	view.MouseMode = tea.MouseModeCellMotion
	return view
}
