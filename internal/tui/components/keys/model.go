// This file deals with application wide keys
package keys

import (
	"charm.land/bubbles/v2/key"
)

// Model contains workspace-wide key bindings.
type Model struct {
	Previous   key.Binding
	Next       key.Binding
	Metrics    key.Binding
	Drag       key.Binding
	Zoom       key.Binding
	ResetGraph key.Binding
	ToggleHelp key.Binding
	Quit       key.Binding
}

// NewModel creates the default workspace key map.
func NewModel() *Model {
	return &Model{
		Previous: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "previous"),
		),
		Next: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next"),
		),
		Metrics: key.NewBinding(
			key.WithKeys("M", "shift+m"),
			key.WithHelp("M", "metrics"),
		),
		Drag: key.NewBinding(
			key.WithKeys("drag"),
			key.WithHelp("drag", "move graph"),
		),
		Zoom: key.NewBinding(
			key.WithKeys("wheel"),
			key.WithHelp("wheel", "zoom graph"),
		),
		ResetGraph: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "fit graph"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
