package pager

import tea "charm.land/bubbletea/v2"

// Init starts no background work.
func (Model) Init() tea.Cmd { return nil }

// Update handles viewport changes and page clicks.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(0, msg.Width)
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft || msg.Y != 0 || m.count == 0 {
			break
		}
		start := max(0, (m.width-indicatorWidth(m.count))/2)
		for index := range m.count {
			if distance := msg.X - (start + index*spacing); distance >= -1 && distance <= 1 {
				m.selected = index
				break
			}
		}
	}
	return m, nil
}
