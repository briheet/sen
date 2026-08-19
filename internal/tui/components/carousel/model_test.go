package carousel

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestCarouselNavigationAndOverflow(t *testing.T) {
	m := NewModel([]string{"api", "worker", "cache"}, WithOverflowIndicators())

	m, _ = m.Update(tea.WindowSizeMsg{Width: 10, Height: 1})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})

	require.Equal(t, 1, m.Selected())
	require.LessOrEqual(t, lipgloss.Width(m.View()), 10)
}

func TestCarouselComposesMultilineItems(t *testing.T) {
	styles := DefaultStyles()
	styles.Item = styles.Item.Border(lipgloss.RoundedBorder(), true)
	styles.Selected = styles.Selected.Border(lipgloss.RoundedBorder(), true)
	m := NewModel([]string{"api", "worker"}, WithOverflowIndicators(), WithStyles(styles))

	m, _ = m.Update(tea.WindowSizeMsg{Width: 12})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})

	require.Equal(t, 3, lipgloss.Height(m.View()))
	require.LessOrEqual(t, lipgloss.Width(m.View()), 12)
}

func TestCarouselSelectsVisibleItemWithMouse(t *testing.T) {
	m := NewModel([]string{"api", "worker", "cache"})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 30, Height: 1})

	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 0, Button: tea.MouseLeft})

	require.Equal(t, 1, m.Selected())
}
