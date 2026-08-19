package carousel

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// View renders the visible carousel items.
func (m Model) View() string {
	if len(m.items) == 0 {
		return ""
	}

	separator := " "
	if m.showSeparators {
		separator = m.styles.Separator.Render(" " + m.separator + " ")
	}
	render := func(start, end int) string {
		items := make([]string, 0, end-start)
		for index := start; index < end; index++ {
			style := m.styles.Item
			if index == m.cursor {
				style = m.styles.Selected
			}
			items = append(items, style.Render(m.items[index]))
		}
		return strings.Join(items, separator)
	}

	content := render(0, len(m.items))
	if m.width <= 0 || lipgloss.Width(content) <= m.width {
		return content
	}

	start, end := 0, m.cursor+1
	left := m.styles.OverflowIndicator.Render(m.leftOverflowIndicator + " ")
	right := m.styles.OverflowIndicator.Render(" " + m.rightOverflowIndicator)
	for start < m.cursor {
		candidate := render(start, end)
		if m.showOverflowIndicators && start > 0 {
			candidate = left + candidate
		}
		if m.showOverflowIndicators && end < len(m.items) {
			candidate += right
		}
		if lipgloss.Width(candidate) <= m.width {
			break
		}
		start++
	}
	for end < len(m.items) {
		candidate := render(start, end+1)
		if m.showOverflowIndicators && start > 0 {
			candidate = left + candidate
		}
		if m.showOverflowIndicators && end+1 < len(m.items) {
			candidate += right
		}
		if lipgloss.Width(candidate) > m.width {
			break
		}
		end++
	}

	content = render(start, end)
	if m.showOverflowIndicators && start > 0 {
		content = left + content
	}
	if m.showOverflowIndicators && end < len(m.items) {
		content += right
	}
	style := lipgloss.NewStyle().MaxWidth(m.width)
	if m.height > 0 {
		style = style.Height(m.height)
	}
	return style.Render(content)
}
