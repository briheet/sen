package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components"
	"github.com/davecgh/go-spew/spew"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
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

	previousFooterHeight := lipgloss.Height(m.footer.View())
	previousPage := m.ctx.ActivePage()
	var headerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(componentMsg)
	var footerCmd tea.Cmd
	m.footer, footerCmd = m.footer.Update(componentMsg)
	if page, ok := m.ctx.Page(m.ctx.ActivePage()); ok {
		headerHeight := lipgloss.Height(m.header.View())
		footerHeight := lipgloss.Height(m.footer.View())
		pageMsg := components.OffsetMouse(componentMsg, 1, 1+headerHeight)
		pageChanged := previousPage != m.ctx.ActivePage()
		if _, resized := componentMsg.(tea.WindowSizeMsg); resized || previousFooterHeight != footerHeight || pageChanged {
			// Newly selected pages still need the current viewport dimensions.
			pageMsg = tea.WindowSizeMsg{
				Width:  max(0, m.width-2),
				Height: max(0, m.height-2-headerHeight-footerHeight),
			}
		}
		page, pageCmd := page.Update(pageMsg)
		m.ctx.SetPage(page)
		return m, tea.Batch(headerCmd, footerCmd, pageCmd)
	}
	return m, tea.Batch(headerCmd, footerCmd)
}
