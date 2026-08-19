package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	return m.view
}

// refreshView rebuilds terminal text when visible component state changes.
func (m *model) refreshView() {
	var view tea.View
	if page, ok := m.ctx.Page(m.ctx.ActivePage()); ok {
		view = page.View()
	}

	width := max(0, m.width)
	bodyHeight := max(0, m.height-1)
	body := lipgloss.NewStyle().Width(width).Height(bodyHeight).Render(view.Content)
	if panel := m.statusbar.HelpView(); panel != "" {
		canvas := lipgloss.NewCanvas(width, bodyHeight)
		canvas.Compose(lipgloss.NewCompositor(
			lipgloss.NewLayer(body),
			lipgloss.NewLayer(panel).
				X(max(0, (width-lipgloss.Width(panel))/2)).
				Y(max(0, (bodyHeight-lipgloss.Height(panel))/2)),
		))
		body = canvas.Render()
	}
	bar := m.statusbar.View()

	// Pages own the remaining viewport; Kitty pixels are placed behind this text.
	layout := lipgloss.JoinVertical(lipgloss.Left, body, bar)
	view.Content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(layout)
	view.AltScreen = true
	view.WindowTitle = "sen"
	m.view = view
}
