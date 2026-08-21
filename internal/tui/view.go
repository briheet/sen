package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/pages"
)

func (m *model) View() tea.View {
	return m.view
}

// refreshView rebuilds terminal text when visible component state changes.
func (m *model) refreshView() {
	width := max(0, m.width)
	bodyHeight := max(0, m.height-1)
	pageName := m.ctx.ActivePage()
	if page, ok := m.ctx.Page(pageName); ok {
		revision := page.Revision()
		if m.bodyPage != pageName || m.bodyRevision != revision || m.bodyWidth != width ||
			m.bodyHeight != bodyHeight || m.bodyEpoch != m.renderEpoch {
			m.body = page.View()
			m.body.Content = fitBody(m.body.Content, width, bodyHeight)
			m.bodyPage = pageName
			m.bodyRevision = revision
			m.bodyWidth = width
			m.bodyHeight = bodyHeight
			m.bodyEpoch = m.renderEpoch
		}
	} else {
		m.body = tea.View{}
	}
	body := m.body.Content
	if panel := m.statusbar.HelpView(); panel != "" {
		body = pages.Overlay(&m.canvas, width, bodyHeight, body, panel)
	}
	bar := m.statusbar.View()

	// Pages own the remaining viewport; Kitty pixels are placed behind this text.
	var content strings.Builder
	content.Grow(len(body) + len(bar) + 1)
	content.WriteString(body)
	if bodyHeight > 0 {
		content.WriteByte('\n')
	}
	if m.height > 0 {
		content.WriteString(bar)
	}
	view := m.body
	view.Content = content.String()
	view.AltScreen = true
	view.WindowTitle = "sen"
	m.view = view
}

func fitBody(content string, width, height int) string {
	if height == 0 {
		return ""
	}
	if lipgloss.Width(content) == width && lipgloss.Height(content) == height {
		return content
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}
