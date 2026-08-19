package components

import tea "charm.land/bubbletea/v2"

// OffsetMouse translates terminal mouse coordinates while preserving event types.
func OffsetMouse(msg tea.Msg, x, y int) tea.Msg {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		msg.X -= x
		msg.Y -= y
		return msg
	case tea.MouseMotionMsg:
		msg.X -= x
		msg.Y -= y
		return msg
	case tea.MouseReleaseMsg:
		msg.X -= x
		msg.Y -= y
		return msg
	default:
		return msg
	}
}
