// This file deals with server model implementation
package servers

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

// switchGraphMsg selects a graph after the previous Kitty image is removed.
type switchGraphMsg struct {
	service string
	page    int
}

// Init starts commands required by the server page.
func (m Model) Init() tea.Cmd {
	commands := make([]tea.Cmd, 0, len(m.graphs))
	for index := range m.graphs {
		if command := m.graphs[index].Init(); command != nil {
			commands = append(commands, command)
		}
	}
	return tea.Batch(commands...)
}

// Update handles messages for the server page.
func (m Model) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if change, ok := msg.(switchGraphMsg); ok {
		if change.service != m.Name() || change.page != m.pending {
			return m, nil
		}
		m.pending = -1
		m.pager.Page = change.page
		m.revision++
		viewport := m.viewport
		viewport.Width = m.width
		viewport.Height = max(0, m.height-1)
		viewport.Visible = m.viewport.Visible
		var command tea.Cmd
		m.graphs[change.page], command = m.graphs[change.page].Update(viewport)
		return m, command
	}

	if viewport, ok := msg.(pages.ViewportMsg); ok {
		m.viewport = viewport
		m.width = max(0, viewport.Width)
		m.height = max(0, viewport.Height)
		commands := make([]tea.Cmd, 0, len(m.graphs))
		for index := range m.graphs {
			graphViewport := viewport
			graphViewport.Width = m.width
			graphViewport.Height = max(0, m.height-1)
			graphViewport.Visible = viewport.Visible && index == m.pager.Page && m.pending < 0
			var command tea.Cmd
			m.graphs[index], command = m.graphs[index].Update(graphViewport)
			if command != nil {
				commands = append(commands, command)
			}
		}
		return m, tea.Batch(commands...)
	}

	if click, ok := msg.(tea.MouseClickMsg); ok && click.Y == m.height-1 {
		indicator := m.pager.View()
		start := max(0, (m.width-lipgloss.Width(indicator))/2)
		if offset := click.X - start; click.Button == tea.MouseLeft && offset >= 0 && offset < lipgloss.Width(indicator) {
			page := min(m.pager.TotalPages-1, offset/lipgloss.Width(m.pager.ActiveDot))
			if page != m.pager.Page && m.pending < 0 {
				previous := m.pager.Page
				m.pending = page
				viewport := m.viewport
				viewport.Width = m.width
				viewport.Height = max(0, m.height-1)

				hidden := viewport
				hidden.Visible = false
				var hideCommand tea.Cmd
				m.graphs[previous], hideCommand = m.graphs[previous].Update(hidden)

				showCommand := func() tea.Msg {
					return switchGraphMsg{service: m.Name(), page: page}
				}
				// Select the next page only after deletion, so Bubble Tea paints its
				// labels before the new Kitty image is placed behind them.
				return m, tea.Sequence(hideCommand, showCommand)
			}
		}
		return m, nil
	}
	m.pager, _ = m.pager.Update(msg)
	switch msg.(type) {
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
		var command tea.Cmd
		m.graphs[m.pager.Page], command = m.graphs[m.pager.Page].Update(msg)
		return m, command
	}

	// Internal render messages are owner-tagged, so every graph can safely consume them.
	commands := make([]tea.Cmd, 0, len(m.graphs))
	for index := range m.graphs {
		var command tea.Cmd
		m.graphs[index], command = m.graphs[index].Update(msg)
		if command != nil {
			commands = append(commands, command)
		}
	}
	return m, tea.Batch(commands...)
}
