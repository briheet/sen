package carousel

import "charm.land/lipgloss/v2"

// View renders the visible carousel items.
func (m Model) View() string {
	if len(m.items) == 0 {
		return ""
	}
	start, end := m.visibleRange()
	content := m.render(start, end)
	if m.showOverflowIndicators && start > 0 {
		content = lipgloss.JoinHorizontal(lipgloss.Bottom, m.leftIndicator(), content)
	}
	if m.showOverflowIndicators && end < len(m.items) {
		content = lipgloss.JoinHorizontal(lipgloss.Bottom, content, m.rightIndicator())
	}
	style := lipgloss.NewStyle().MaxWidth(m.width)
	if m.height > 0 {
		style = style.Height(m.height)
	}
	return style.Render(content)
}

func (m Model) render(start, end int) string {
	separator := " "
	if m.showSeparators {
		separator = m.styles.Separator.Render(" " + m.separator + " ")
	}
	items := make([]string, 0, 2*(end-start)-1)
	for index := start; index < end; index++ {
		if index > start {
			items = append(items, separator)
		}
		items = append(items, m.itemView(index))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, items...)
}

func (m Model) visibleRange() (int, int) {
	start, end := 0, len(m.items)
	content := m.render(start, end)
	if m.width <= 0 || lipgloss.Width(content) <= m.width {
		return start, end
	}

	end = m.cursor + 1
	for start < m.cursor {
		candidate := m.render(start, end)
		if m.showOverflowIndicators && start > 0 {
			candidate = lipgloss.JoinHorizontal(lipgloss.Bottom, m.leftIndicator(), candidate)
		}
		if m.showOverflowIndicators && end < len(m.items) {
			candidate = lipgloss.JoinHorizontal(lipgloss.Bottom, candidate, m.rightIndicator())
		}
		if lipgloss.Width(candidate) <= m.width {
			break
		}
		start++
	}
	for end < len(m.items) {
		candidate := m.render(start, end+1)
		if m.showOverflowIndicators && start > 0 {
			candidate = lipgloss.JoinHorizontal(lipgloss.Bottom, m.leftIndicator(), candidate)
		}
		if m.showOverflowIndicators && end+1 < len(m.items) {
			candidate = lipgloss.JoinHorizontal(lipgloss.Bottom, candidate, m.rightIndicator())
		}
		if lipgloss.Width(candidate) > m.width {
			break
		}
		end++
	}
	return start, end
}

func (m Model) itemAt(x int) int {
	start, end := m.visibleRange()
	if m.showOverflowIndicators && start > 0 {
		x -= lipgloss.Width(m.leftIndicator())
	}
	for index := start; index < end; index++ {
		width := lipgloss.Width(m.itemView(index))
		if x >= 0 && x < width {
			return index
		}
		x -= width
		if index < end-1 {
			separatorWidth := 1
			if m.showSeparators {
				separatorWidth = lipgloss.Width(m.styles.Separator.Render(" " + m.separator + " "))
			}
			x -= separatorWidth
		}
	}
	return -1
}

func (m Model) itemView(index int) string {
	style := m.styles.Item
	if index == m.cursor {
		style = m.styles.Selected
	}
	return style.Render(m.items[index])
}

func (m Model) leftIndicator() string {
	return m.styles.OverflowIndicator.Render(m.leftOverflowIndicator + " ")
}

func (m Model) rightIndicator() string {
	return m.styles.OverflowIndicator.Render(" " + m.rightOverflowIndicator)
}
