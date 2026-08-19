package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestOffsetMousePreservesEventType(t *testing.T) {
	message := OffsetMouse(tea.MouseMotionMsg{X: 12, Y: 8, Button: tea.MouseLeft}, 2, 3)

	motion, ok := message.(tea.MouseMotionMsg)
	require.True(t, ok)
	require.Equal(t, 10, motion.X)
	require.Equal(t, 5, motion.Y)
	require.Equal(t, tea.MouseLeft, motion.Button)
}
