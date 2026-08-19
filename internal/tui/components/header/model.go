// Package header contains the workspace header.
package header

import (
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components/carousel"
	"github.com/briheet/sen/internal/tui/components/tabs"
	"github.com/briheet/sen/internal/tui/context"
	"github.com/briheet/sen/internal/tui/styles"
)

const horizontalPadding = 1

// Model contains workspace tabs and their navigation state.
type Model struct {
	tabs []tabs.Model // tabs positioned on the header

	carousel carousel.Model          // tabs state
	ctx      *context.ProgramContext // pointer to program context

	style Style // Header style
	width int

	name    string // application's name
	version string // application's version
}

// NewModel creates tabs from the configured service pages.
func NewModel(ctx *context.ProgramContext) Model {
	pageModels := ctx.Pages()
	tabModels := make([]tabs.Model, 0, len(pageModels))
	items := make([]string, 0, len(pageModels))
	for _, page := range pageModels {
		tab := tabs.NewModel(page.Name())
		tabModels = append(tabModels, tab)
		items = append(items, tab.View())
	}
	c := carousel.NewModel(items,
		carousel.WithOverflowIndicators("←", "→"),
		carousel.WithSeparators(),
	)

	return Model{
		tabs:     tabModels,
		carousel: c,
		ctx:      ctx,
		style:    DefaultStyle(),
		name:     "sen",
		version:  "0.1.0",
	}
}

// Style contains header presentation styles.
type Style struct {
	Root lipgloss.Style
	Logo lipgloss.Style
}

// DefaultStyle returns header styles using the Zakura theme.
func DefaultStyle() Style {
	return Style{
		Root: lipgloss.NewStyle().
			Foreground(styles.Zakura.Text).
			Padding(0, horizontalPadding),
		Logo: lipgloss.NewStyle().
			Foreground(styles.Zakura.Primary).
			Bold(true).
			Padding(0, 1),
	}
}
