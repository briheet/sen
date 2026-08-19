// This file deals with server model implementation
package servers

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

// Init starts commands required by the server page.
func (m Model) Init() tea.Cmd {
	return m.graph.Init()
}

// Update handles messages for the server page.
func (m Model) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if viewport, ok := msg.(pages.ViewportMsg); ok {
		m.viewport = viewport
		m.width = max(0, viewport.Width)
		m.height = max(0, viewport.Height)
		viewport.Width = m.width
		viewport.Height = max(0, m.height-1)
		viewport.Visible = viewport.Visible && m.pager.Page == 0
		var command tea.Cmd
		m.graph, command = m.graph.Update(viewport)
		return m, command
	}

	if click, ok := msg.(tea.MouseClickMsg); ok && click.Y == m.height-1 {
		indicator := m.pager.View()
		start := max(0, (m.width-lipgloss.Width(indicator))/2)
		if offset := click.X - start; click.Button == tea.MouseLeft && offset >= 0 && offset < lipgloss.Width(indicator) {
			page := min(m.pager.TotalPages-1, offset/lipgloss.Width(m.pager.ActiveDot))
			if page != m.pager.Page {
				m.pager.Page = page
				m.revision++
				viewport := m.viewport
				viewport.Height = max(0, m.height-1)
				viewport.Visible = m.viewport.Visible && page == 0
				var command tea.Cmd
				m.graph, command = m.graph.Update(viewport)
				return m, command
			}
		}
		return m, nil
	}
	if m.pager.Page != 0 {
		switch msg.(type) {
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	m.pager, _ = m.pager.Update(msg)
	var command tea.Cmd
	m.graph, command = m.graph.Update(msg)
	return m, command
}
