package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case enginesDoneMsg:
		return m, tea.Quit
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	componentMsg := msg
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
		componentMsg = tea.WindowSizeMsg{
			Width:  max(0, size.Width-2),
			Height: max(0, size.Height-2),
		}
	}

	var headerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(componentMsg)
	var footerCmd tea.Cmd
	m.footer, footerCmd = m.footer.Update(componentMsg)
	if page, ok := m.ctx.Page(m.ctx.ActivePage()); ok {
		page, pageCmd := page.Update(componentMsg)
		m.ctx.SetPage(page)
		return m, tea.Batch(headerCmd, footerCmd, pageCmd)
	}
	return m, tea.Batch(headerCmd, footerCmd)
}
