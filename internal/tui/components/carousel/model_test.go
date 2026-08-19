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
