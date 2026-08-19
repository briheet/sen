package pager

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/stretchr/testify/require"
)

func TestPagerSelectsClickedDot(t *testing.T) {
	pager, _ := New(3, styles.Zakura).Update(tea.WindowSizeMsg{Width: 21, Height: 1})
	pager, _ = pager.Update(tea.MouseClickMsg{X: 13, Y: 0, Button: tea.MouseLeft})

	require.Equal(t, 2, pager.Selected())
	require.Equal(t, 21, lipgloss.Width(pager.View()))
}
