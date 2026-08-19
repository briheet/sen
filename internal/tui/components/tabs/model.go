package tabs

// Model is one service tab in the workspace header.
type Model struct {
	name string
}

// NewModel creates a tab for a service page.
func NewModel(name string) Model {
	return Model{name: name}
}

// Name returns the page selected by this tab.
func (m Model) Name() string {
	return m.name
}
