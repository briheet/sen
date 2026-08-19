package header

import "charm.land/lipgloss/v2"

// View renders tabs on the left and the application identity on the right.
func (m Model) View() string {
	tabs := m.carousel.View()
	logo := m.style.Logo.Render(m.name + " " + m.version)
	tabs = lipgloss.NewStyle().
		Width(max(0, m.width-lipgloss.Width(logo)-2*horizontalPadding)).
		Render(tabs)
	return m.style.Root.
		Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, logo))
}
