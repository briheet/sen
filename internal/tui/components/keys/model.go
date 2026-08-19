// This file deals with application wide keys
package keys

import (
	"charm.land/bubbles/v2/key"
)

// Model contains workspace-wide key bindings.
type Model struct {
	Previous   key.Binding
	Next       key.Binding
	ToggleHelp key.Binding
	Quit       key.Binding
}

// NewModel creates the default workspace key map.
func NewModel() *Model {
	return &Model{
		Previous: key.NewBinding(
			key.WithKeys("left", "shift+h"),
			key.WithHelp("←/h", "previous"),
		),
		Next: key.NewBinding(
			key.WithKeys("right", "shift+l"),
			key.WithHelp("→/l", "next"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
