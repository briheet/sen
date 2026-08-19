package statusbar

import (
	"charm.land/lipgloss/v2"
)

// View renders project identity, service tabs, and the help affordance.
func (m Model) View() string {
	left, right := m.leftView(), m.rightView()
	services := lipgloss.NewStyle().
		Width(m.serviceWidth()).
		MaxWidth(m.serviceWidth()).
		Render(m.carousel.View())
	return m.style.Root.Width(m.width).MaxWidth(m.width).Render(left + services + right)
}

// HelpView renders the centered help panel content.
func (m Model) HelpView() string {
	if !m.showHelp || m.helpWidth() < 20 {
		return ""
	}
	content := m.style.HelpTitle.Render("HELP") + "\n" + m.help.View(m.keys)
	return m.style.HelpPanel.Width(m.helpWidth()).Render(content)
}

func (m Model) leftView() string {
	return m.style.Brand.Render("sen") + m.style.Separator.Render("│") + " "
}

func (m Model) rightView() string {
	return " " + m.style.Separator.Render("│") + m.style.HelpPrompt.Render("? help")
}

func (m Model) serviceWidth() int {
	return max(0, m.width-lipgloss.Width(m.leftView())-lipgloss.Width(m.rightView()))
}

func (m Model) helpWidth() int {
	return min(48, max(0, m.width-8))
}
