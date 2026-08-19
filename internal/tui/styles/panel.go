package styles

import "charm.land/lipgloss/v2"

// Panel returns the shared rounded style used by modal surfaces.
func Panel(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.TextMuted).
		Padding(0, 1)
}
