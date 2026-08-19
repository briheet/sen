package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/davecgh/go-spew/spew"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		if summary, ok := msg.(interface{ DebugSummary() string }); ok {
			_, _ = fmt.Fprintln(m.dump, summary.DebugSummary())
		} else if _, raw := msg.(tea.RawMsg); raw {
			// Raw messages contain the full encoded graph image.
			_, _ = fmt.Fprintln(m.dump, "(tea.RawMsg) image payload omitted")
		} else {
			spew.Fdump(m.dump, msg)
		}
	}
	switch msg := msg.(type) {
	case enginesDoneMsg:
		return m, tea.Quit
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}
	refresh := false
	switch msg.(type) {
	case tea.WindowSizeMsg:
		refresh = true
	case tea.KeyPressMsg, tea.MouseClickMsg, tea.ColorProfileMsg, tea.BackgroundColorMsg:
		refresh = true
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
	previousPageName := m.ctx.ActivePage()
	var headerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(componentMsg)
	var footerCmd tea.Cmd
	m.footer, footerCmd = m.footer.Update(componentMsg)
	headerHeight := lipgloss.Height(m.header.View())
	footerHeight := lipgloss.Height(m.footer.View())
	viewport := pages.ViewportMsg{
		X: 1, Y: 1 + headerHeight,
		Width:   max(0, m.width-2),
		Height:  max(0, m.height-2-headerHeight-footerHeight),
		Visible: true,
	}
	activePageName := m.ctx.ActivePage()
	pageChanged := previousPageName != activePageName
	var hiddenCmd tea.Cmd
	if pageChanged {
		refresh = true
		if previousPage, ok := m.ctx.Page(previousPageName); ok {
			hidden := viewport
			hidden.Visible = false
			previousPage, hiddenCmd = previousPage.Update(hidden)
			m.ctx.SetPage(previousPage)
		}
	}
	if page, ok := m.ctx.Page(activePageName); ok {
		pageMsg := components.OffsetMouse(componentMsg, viewport.X, viewport.Y)
		if _, resized := componentMsg.(tea.WindowSizeMsg); resized || previousFooterHeight != footerHeight || pageChanged {
			pageMsg = viewport
		}
		previousRevision := page.Revision()
		page, pageCmd := page.Update(pageMsg)
		m.ctx.SetPage(page)
		if page.Revision() != previousRevision {
			refresh = true
		}
		if refresh {
			m.refreshView()
		}
		return m, tea.Batch(headerCmd, footerCmd, tea.Sequence(hiddenCmd, pageCmd))
	}
	if refresh {
		m.refreshView()
	}
	return m, tea.Batch(headerCmd, footerCmd)
}
