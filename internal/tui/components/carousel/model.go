package carousel

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model manages horizontal item selection and overflow.
type Model struct {
	KeyMap KeyMap

	cursor                 int
	showSeparators         bool
	separator              string
	showOverflowIndicators bool
	leftOverflowIndicator  string
	rightOverflowIndicator string
	styles                 Style

	width  int
	height int

	focus bool

	items []string
}

// KeyMap defines carousel navigation keys.
type KeyMap struct {
	SelectLeft  key.Binding
	SelectRight key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		SelectLeft: key.NewBinding(
			key.WithKeys("left", "shift+h"),
			key.WithHelp("←/H", "previous"),
		),
		SelectRight: key.NewBinding(
			key.WithKeys("right", "shift+l"),
			key.WithHelp("→/L", "next"),
		),
	}
}

// Style contains carousel presentation styles.
type Style struct {
	Item              lipgloss.Style
	Selected          lipgloss.Style
	OverflowIndicator lipgloss.Style
	Separator         lipgloss.Style
}

// DefaultStyles returns styles derived from the default theme.
func DefaultStyles() Style {
	return Style{
		Item: lipgloss.NewStyle().
			Foreground(styles.Zakura.TextMuted),
		Selected: lipgloss.NewStyle().
			Foreground(styles.Zakura.NodeActive).
			Bold(true).
			BorderBottom(true).
			BorderForeground(styles.Zakura.NodeHot),
		OverflowIndicator: lipgloss.NewStyle().
			Foreground(styles.Zakura.Border),
		Separator: lipgloss.NewStyle().
			Foreground(styles.Zakura.Border),
	}
}

// Option configures a carousel model.
type Option func(*Model)

// NewModel creates a carousel model with its initial items.
func NewModel(items []string, opts ...Option) Model {
	m := Model{
		cursor: 0,
		focus:  true,
		items:  append([]string(nil), items...),

		KeyMap:                 DefaultKeyMap(),
		styles:                 DefaultStyles(),
		leftOverflowIndicator:  "<",
		rightOverflowIndicator: ">",
		separator:              "|",
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// WithOverflowIndicators enables indicators for hidden items.
func WithOverflowIndicators(indicators ...string) Option {
	return func(m *Model) {
		m.showOverflowIndicators = true
		if len(indicators) > 0 {
			m.leftOverflowIndicator = indicators[0]
		}
		if len(indicators) > 1 {
			m.rightOverflowIndicator = indicators[1]
		}
	}
}

// WithSeparators enables separators between items.
func WithSeparators(sep ...string) Option {
	return func(m *Model) {
		m.showSeparators = true
		if len(sep) > 0 {
			m.separator = sep[0]
		}
	}
}

// Selected returns the selected item index.
func (m Model) Selected() int {
	return m.cursor
}
