package statusbar

import (
	"charm.land/lipgloss/v2"
)

// View renders project identity, service tabs, and the help affordance.
func (m *Model) View() string {
	if m.viewValid {
		return m.view
	}
	left, right := m.leftView(), m.rightView()
	services := lipgloss.NewStyle().
		Width(m.serviceWidth()).
		MaxWidth(m.serviceWidth()).
		Render(m.carousel.View())
	m.view = m.style.Root.Width(m.width).MaxWidth(m.width).Render(left + services + right)
	m.viewValid = true
	return m.view
}

// HelpView renders the centered help panel content.
func (m *Model) HelpView() string {
	if m.helpViewValid {
		return m.helpView
	}
	if !m.showHelp || m.helpWidth() < 20 {
		m.helpView = ""
		m.helpViewValid = true
		return ""
	}
	content := m.style.HelpTitle.Render("HELP") + "\n" + m.help.View(m.keys)
	m.helpView = m.style.HelpPanel.Width(m.helpWidth()).Render(content)
	m.helpViewValid = true
	return m.helpView
}

func (m *Model) leftView() string {
	return m.style.Brand.Render("sen") + m.style.Separator.Render("│") + " "
}

func (m *Model) rightView() string {
	return " " + m.style.Separator.Render("│") + m.style.HelpPrompt.Render("? help")
}

func (m *Model) serviceWidth() int {
	return max(0, m.width-lipgloss.Width(m.leftView())-lipgloss.Width(m.rightView()))
}

func (m *Model) helpWidth() int {
	return min(48, max(0, m.width-8))
}
