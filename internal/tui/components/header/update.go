package header

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Init starts commands required by the header.
func (m Model) Init() tea.Cmd {
	commands := make([]tea.Cmd, 0, len(m.tabs)+1)
	commands = append(commands, m.carousel.Init())
	for _, tab := range m.tabs {
		commands = append(commands, tab.Init())
	}
	return tea.Batch(commands...)
}

// Update handles header navigation and terminal resizing.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		logo := m.style.Logo.Render(m.name + " " + m.version)
		msg = tea.WindowSizeMsg{
			Width:  max(0, size.Width-lipgloss.Width(logo)-2*horizontalPadding),
			Height: 1,
		}
	}

	var cmd tea.Cmd
	m.carousel, cmd = m.carousel.Update(msg)
	if len(m.tabs) != 0 {
		m.ctx.SelectPage(m.tabs[m.carousel.Selected()].Name())
	}
	return m, cmd
}
