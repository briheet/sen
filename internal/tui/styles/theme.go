package styles

import (
	"image/color"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
)

type Theme struct {
	Primary   color.Color
	Secondary color.Color

	Text      color.Color
	TextMuted color.Color
	Border    color.Color

	NodeIdle   color.Color
	NodeActive color.Color
	NodeHot    color.Color

	CPU    color.Color
	Memory color.Color

	Success color.Color
	Warning color.Color
	Error   color.Color
}

// Zakura is the default theme inspired by Senbonzakura (千本桜).
var Zakura = Theme{
	// Signature Senbonzakura petals
	Primary:   lipgloss.Color("#F5A9D0"), // luminous sakura
	Secondary: lipgloss.Color("#D889B7"), // deeper petal pink

	// Byakuya's white/silver clothing against the darkness
	Text:      lipgloss.Color("#F2EDF3"), // silver-white
	TextMuted: lipgloss.Color("#B77F9E"), // muted dusty sakura
	Border:    lipgloss.Color("#514B50"), // ash-grey / dark plum

	// Runtime graph — petals becoming more intense
	NodeIdle:   lipgloss.Color("#625462"), // dormant blade
	NodeActive: lipgloss.Color("#F3A6D2"), // scattered petals
	NodeHot:    lipgloss.Color("#FF6FB5"), // concentrated blades

	// Profiling
	CPU:    lipgloss.Color("#D0A765"), // subtle bronze / sword guard
	Memory: lipgloss.Color("#B4A0D6"), // lavender

	// State
	Success: lipgloss.Color("#9EB59F"),
	Warning: lipgloss.Color("#D0A765"),
	Error:   lipgloss.Color("#D86B82"),
}

// Cli ColorScheme function
func FangColorScheme(theme Theme) fang.ColorSchemeFunc {
	return func(_ lipgloss.LightDarkFunc) fang.ColorScheme {
		return fang.ColorScheme{
			// General text / root description
			Base:        theme.TextMuted,
			Title:       theme.Primary,
			Description: theme.TextMuted,

			// Usage / example blocks — dark ash-grey
			Codeblock: lipgloss.Color("#151416"),

			// Executable / commands
			Program: theme.Primary,
			Command: theme.Secondary,

			// Arguments
			Argument:       theme.Secondary,
			DimmedArgument: theme.Secondary,
			QuotedString:   theme.Secondary,

			// Flags
			Flag:        theme.Secondary,
			FlagDefault: lipgloss.Color("#81747C"),

			// Supporting text
			Dash:    theme.Secondary,
			Help:    lipgloss.Color("#887582"),
			Comment: lipgloss.Color("#887582"),

			// Errors
			ErrorHeader: [2]color.Color{
				theme.Text,
				theme.Error,
			},
			ErrorDetails: theme.Error,
		}
	}
}
