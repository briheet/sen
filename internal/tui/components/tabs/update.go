package tabs

import tea "charm.land/bubbletea/v2"

// Init starts commands required by the tab.
func (Model) Init() tea.Cmd {
	return nil
}

// Update handles tab messages.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}
