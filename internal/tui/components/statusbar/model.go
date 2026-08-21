// Package statusbar provides workspace navigation and contextual help.
package statusbar

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components/carousel"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/context"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model contains the bottom workspace navigation state.
type Model struct {
	carousel      carousel.Model
	help          help.Model
	keys          *keys.Model
	ctx           *context.ProgramContext
	style         Style
	width         int
	showHelp      bool
	view          string
	helpView      string
	viewValid     bool
	helpViewValid bool
}

// New creates a status bar for the configured service pages.
func New(ctx *context.ProgramContext, keys *keys.Model, theme styles.Theme) Model {
	pageModels := ctx.Pages()
	names := make([]string, 0, len(pageModels))
	for _, page := range pageModels {
		names = append(names, page.Name())
	}
	style := defaultStyle(theme)
	helpModel := help.New()
	helpModel.ShowAll = true
	helpModel.Styles = style.Help
	return Model{
		carousel: carousel.NewModel(names,
			carousel.WithOverflowIndicators("←", "→"),
			carousel.WithStyles(style.Services),
		),
		help:  helpModel,
		keys:  keys,
		ctx:   ctx,
		style: style,
	}
}

// Style contains status bar and help panel presentation.
type Style struct {
	Root       lipgloss.Style
	Brand      lipgloss.Style
	Separator  lipgloss.Style
	HelpPrompt lipgloss.Style
	HelpPanel  lipgloss.Style
	HelpTitle  lipgloss.Style
	Services   carousel.Style
	Help       help.Styles
}

func defaultStyle(theme styles.Theme) Style {
	key := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	description := lipgloss.NewStyle().Foreground(theme.TextMuted)
	separator := lipgloss.NewStyle().Foreground(theme.Border)
	return Style{
		Root:       lipgloss.NewStyle().Foreground(theme.TextMuted),
		Brand:      lipgloss.NewStyle().Foreground(theme.Primary).Bold(true).Padding(0, 1),
		Separator:  separator,
		HelpPrompt: lipgloss.NewStyle().Foreground(theme.Primary).Bold(true).Padding(0, 1),
		HelpPanel:  styles.Panel(theme),
		HelpTitle:  key,
		Services: carousel.Style{
			Item:              lipgloss.NewStyle().Foreground(theme.TextMuted).Padding(0, 1),
			Selected:          lipgloss.NewStyle().Foreground(theme.Primary).Reverse(true).Bold(true).Padding(0, 1),
			OverflowIndicator: separator,
		},
		Help: help.Styles{
			Ellipsis:       separator,
			ShortKey:       key,
			ShortDesc:      description,
			ShortSeparator: separator,
			FullKey:        key,
			FullDesc:       description,
			FullSeparator:  separator,
		},
	}
}
