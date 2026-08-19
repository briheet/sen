package footer

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Init starts commands required by the footer.
func (Model) Init() tea.Cmd {
	return nil
}

// Update handles help toggling and terminal resizing.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.SetWidth(msg.Width)
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.ToggleHelp) {
			m.help.ShowAll = !m.help.ShowAll
			if m.help.ShowAll {
				m.keys.ToggleHelp.SetHelp("?", "less")
			} else {
				m.keys.ToggleHelp.SetHelp("?", "more")
			}
		}
	}
	m.help, _ = m.help.Update(msg)
	return m, nil
}
