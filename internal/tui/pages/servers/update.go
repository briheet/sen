// This file deals with server model implementation
package servers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/components"
	"github.com/briheet/sen/internal/tui/pages"
)

// Init starts commands required by the server page.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.graph.Init(), m.pager.Init())
}

// Update handles messages for the server page.
func (m Model) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = max(0, size.Width)
		m.height = max(0, size.Height)
		m.pager, _ = m.pager.Update(tea.WindowSizeMsg{Width: m.width, Height: 1})
		var command tea.Cmd
		m.graph, command = m.graph.Update(tea.WindowSizeMsg{Width: m.width, Height: max(0, m.height-1)})
		return m, command
	}

	if click, ok := msg.(tea.MouseClickMsg); ok && click.Y == m.height-1 {
		m.pager, _ = m.pager.Update(components.OffsetMouse(click, 0, m.height-1))
		return m, nil
	}
	if m.pager.Selected() != 0 {
		switch msg.(type) {
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	var command tea.Cmd
	m.graph, command = m.graph.Update(msg)
	return m, command
}
