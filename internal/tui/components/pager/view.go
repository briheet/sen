package pager

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// View centers a filled active dot among hollow inactive dots.
func (m Model) View() string {
	if m.count == 0 {
		return ""
	}
	dots := make([]string, m.count)
	for index := range dots {
		dots[index] = m.inactive.Render("○")
		if index == m.selected {
			dots[index] = m.active.Render("●")
		}
	}
	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(strings.Join(dots, "  "))
}

func indicatorWidth(count int) int {
	if count == 0 {
		return 0
	}
	return 1 + (count-1)*spacing
}
