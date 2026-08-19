package carousel

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Init starts commands required by the carousel.
func (Model) Init() tea.Cmd {
	return nil
}

// Update moves the selection within the carousel.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(0, msg.Width)
		m.height = max(0, msg.Height)
	case tea.KeyPressMsg:
		if !m.focus || len(m.items) == 0 {
			break
		}
		switch {
		case key.Matches(msg, m.KeyMap.SelectLeft):
			m.cursor = max(0, m.cursor-1)
		case key.Matches(msg, m.KeyMap.SelectRight):
			m.cursor = min(len(m.items)-1, m.cursor+1)
		}
	}
	return m, nil
}
