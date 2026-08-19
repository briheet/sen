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

	width := max(0, m.width-2)
	border := lipgloss.RoundedBorder()
	header := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(m.header.View())
	footer := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(m.footer.View())

	// Pages own the remaining viewport; Kitty pixels are placed behind this text.
	layout := lipgloss.JoinVertical(lipgloss.Left, header, view.Content, footer)
	view.Content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Border(border).
		BorderForeground(m.activeTheme.Border).
		Render(layout)
	view.AltScreen = true
	view.WindowTitle = "sen"
	m.view = view
}
