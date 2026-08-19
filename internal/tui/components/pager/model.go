// Package pager renders a compact clickable page indicator.
package pager

import (
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
)

const spacing = 3

// Model owns the selected page and indicator layout.
type Model struct {
	active   lipgloss.Style
	inactive lipgloss.Style
	count    int
	selected int
	width    int
}

// New creates an indicator with the first page selected.
func New(count int, theme styles.Theme) Model {
	return Model{
		active:   lipgloss.NewStyle().Foreground(theme.NodeActive),
		inactive: lipgloss.NewStyle().Foreground(theme.Border),
		count:    max(0, count),
	}
}

// Selected returns the active page index.
func (m Model) Selected() int { return m.selected }
