package statusbar

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components"
)

// Init starts commands required by the service carousel.
func (m Model) Init() tea.Cmd { return m.carousel.Init() }

// Update handles service navigation, resizing, and the help panel.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	selected := m.carousel.Selected()
	carouselMsg := msg
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewValid = false
		m.helpViewValid = false
		m.width = max(0, msg.Width)
		m.help.SetWidth(max(0, m.helpWidth()-4))
		msg.Width = m.serviceWidth()
		msg.Height = 1
		var command tea.Cmd
		m.carousel, command = m.carousel.Update(msg)
		return m, command
	case tea.BackgroundColorMsg:
		m.viewValid = false
		m.helpViewValid = false
		for _, style := range []*lipgloss.Style{
			&m.style.HelpPanel, &m.style.HelpTitle, &m.style.Help.Ellipsis,
			&m.style.Help.ShortKey, &m.style.Help.ShortDesc, &m.style.Help.ShortSeparator,
			&m.style.Help.FullKey, &m.style.Help.FullDesc, &m.style.Help.FullSeparator,
		} {
			*style = style.Background(msg.Color)
		}
		m.help.Styles = m.style.Help
	case tea.ColorProfileMsg:
		m.viewValid = false
		m.helpViewValid = false
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.ToggleHelp) {
			m.showHelp = !m.showHelp
			m.helpViewValid = false
			return m, nil
		}
		if m.showHelp {
			if msg.Code == tea.KeyEscape {
				m.showHelp = false
				m.helpViewValid = false
			}
			return m, nil
		}
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft || msg.Y != 0 {
			return m, nil
		}
		if msg.X >= m.width-lipgloss.Width(m.rightView()) {
			m.showHelp = !m.showHelp
			m.helpViewValid = false
			return m, nil
		}
		if m.showHelp {
			return m, nil
		}
		start := lipgloss.Width(m.leftView())
		if msg.X < start || msg.X >= start+m.serviceWidth() {
			return m, nil
		}
		carouselMsg = components.OffsetMouse(msg, start, 0)
	}

	var command tea.Cmd
	m.carousel, command = m.carousel.Update(carouselMsg)
	if m.carousel.Selected() != selected {
		m.viewValid = false
	}
	pages := m.ctx.Pages()
	if len(pages) != 0 {
		m.ctx.SelectPage(pages[m.carousel.Selected()].Name())
	}
	return m, command
}

// HelpVisible reports whether the help panel is open.
func (m Model) HelpVisible() bool { return m.showHelp }
