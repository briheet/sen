package kv

import (
	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

// Init starts commands required by the key-value page.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initScreen())
}

// Update handles messages for the key-value page.
func (m Model) Update(_ tea.Msg) (pages.Page, tea.Cmd) {
	return m, nil
}
