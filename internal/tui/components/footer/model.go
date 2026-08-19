// Package footer renders workspace key help.
package footer

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model contains the workspace help state.
type Model struct {
	help help.Model
	keys *keys.Model
}

// NewModel creates a footer using the active theme.
func NewModel(keys *keys.Model, theme styles.Theme) Model {
	model := help.New()
	model.Styles = helpStyles(theme)
	return Model{help: model, keys: keys}
}

func helpStyles(theme styles.Theme) help.Styles {
	key := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	description := lipgloss.NewStyle().Foreground(theme.TextMuted)
	separator := lipgloss.NewStyle().Foreground(theme.Border)
	return help.Styles{
		Ellipsis:       separator,
		ShortKey:       key,
		ShortDesc:      description,
		ShortSeparator: separator,
		FullKey:        key,
		FullDesc:       description,
		FullSeparator:  separator,
	}
}
