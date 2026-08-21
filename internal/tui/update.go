package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/components"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/davecgh/go-spew/spew"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	if tick, ok := msg.(pages.TelemetryTickMsg); ok {
		commands, refresh := m.updateTelemetry(tick)
		if refresh {
			m.refreshView()
		}
		return m, pages.BatchCommands(commands)
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
	case tea.ColorProfileMsg, tea.BackgroundColorMsg:
		m.renderEpoch++
		refresh = true
	}
	componentMsg := msg
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
		componentMsg = tea.WindowSizeMsg{
			Width:  max(0, size.Width),
			Height: max(0, size.Height),
		}
	}

	previousPageName := m.ctx.ActivePage()
	bodyHeight := max(0, m.height-1)
	statusMsg := components.OffsetMouse(componentMsg, 0, bodyHeight)
	if _, resized := componentMsg.(tea.WindowSizeMsg); resized {
		statusMsg = tea.WindowSizeMsg{Width: max(0, m.width), Height: 1}
	}
	wasHelpVisible := m.statusbar.HelpVisible()
	var statusCmd tea.Cmd
	m.statusbar, statusCmd = m.statusbar.Update(statusMsg)
	helpChanged := wasHelpVisible != m.statusbar.HelpVisible()
	viewport := pages.ViewportMsg{
		X: 0, Y: 0,
		Width:   max(0, m.width),
		Height:  bodyHeight,
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
		if helpChanged {
			pageMsg = pages.ObscuredMsg{Obscured: m.statusbar.HelpVisible()}
		} else if _, resized := componentMsg.(tea.WindowSizeMsg); resized || pageChanged {
			pageMsg = viewport
		} else if blocksPageInput(componentMsg, bodyHeight, wasHelpVisible || m.statusbar.HelpVisible()) {
			pageMsg = nil
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
		return m, pages.BatchPair(statusCmd, tea.Sequence(hiddenCmd, pageCmd))
	}
	if refresh {
		m.refreshView()
	}
	return m, statusCmd
}

func (m *model) updateTelemetry(msg pages.TelemetryTickMsg) ([]tea.Cmd, bool) {
	active := m.ctx.ActivePage()
	var commands []tea.Cmd
	refresh := false
	for _, page := range m.ctx.Pages() {
		if msg.Service != "" && page.Name() != msg.Service {
			continue
		}
		previousRevision := page.Revision()
		page, command := page.Update(msg)
		m.ctx.SetPage(page)
		if page.Name() == active && page.Revision() != previousRevision {
			refresh = true
		}
		if command != nil {
			commands = append(commands, command)
		}
	}
	return commands, refresh
}

func blocksPageInput(msg tea.Msg, bodyHeight int, helpVisible bool) bool {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return helpVisible
	case tea.MouseClickMsg:
		return helpVisible || msg.Y >= bodyHeight
	case tea.MouseMotionMsg:
		return helpVisible || msg.Y >= bodyHeight
	case tea.MouseReleaseMsg:
		return helpVisible || msg.Y >= bodyHeight
	case tea.MouseWheelMsg:
		return helpVisible || msg.Y >= bodyHeight
	}
	return false
}
