package keys

import "charm.land/bubbles/v2/key"

// ShortHelp returns the bindings shown in the compact footer.
func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.Previous, m.Next, m.ToggleHelp, m.Quit}
}

// FullHelp returns workspace bindings grouped by purpose.
func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Previous, m.Next},
		{m.ToggleHelp, m.Quit},
	}
}
