package pages

import tea "charm.land/bubbletea/v2"

// BatchCommands batches an already compacted command slice without rescanning it.
func BatchCommands(commands []tea.Cmd) tea.Cmd {
	switch len(commands) {
	case 0:
		return nil
	case 1:
		return commands[0]
	default:
		return func() tea.Msg { return tea.BatchMsg(commands) }
	}
}

// BatchPair avoids allocating command slices when either command is absent.
func BatchPair(first, second tea.Cmd) tea.Cmd {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return func() tea.Msg { return tea.BatchMsg{first, second} }
}
