// This file deals with server model implementation
package servers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

// Init starts commands required by the server page.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initScreen())
}

// Update handles messages for the server page.
func (m Model) Update(_ tea.Msg) (pages.Page, tea.Cmd) {
	return m, nil
}
