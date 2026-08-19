// This file deals with server view implementation
package servers

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// View renders the selected server.
func (m Model) View() tea.View {
	content := m.graphs[m.pager.Page].View()
	indicator := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.pager.View())
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, indicator))
	view.MouseMode = tea.MouseModeCellMotion
	return view
}
