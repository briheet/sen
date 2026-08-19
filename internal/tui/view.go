package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	var view tea.View
	if page, ok := m.ctx.Page(m.ctx.ActivePage()); ok {
		view = page.View()
	}

	width := max(0, m.width-2)
	height := max(0, m.height-2)
	border := lipgloss.RoundedBorder()
	header := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(m.header.View())
	footer := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(m.footer.View())
	content := lipgloss.NewStyle().
		Width(width).
		Height(max(0, height-lipgloss.Height(header)-lipgloss.Height(footer))).
		Align(lipgloss.Left, lipgloss.Top).
		Render(view.Content)

	layout := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
	view.Content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Border(border).
		BorderForeground(m.activeTheme.Border).
		Render(layout)
	view.AltScreen = true
	view.WindowTitle = "sen"
	return view
}
