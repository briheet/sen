package keys

import "charm.land/bubbles/v2/key"

// ShortHelp satisfies the help component's compact key map.
func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.ToggleHelp}
}

// FullHelp groups every workspace and graph binding for the help panel.
func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Previous, m.Next, m.Metrics, m.ToggleHelp},
		{m.Drag, m.Zoom, m.ResetGraph, m.Quit},
	}
}
