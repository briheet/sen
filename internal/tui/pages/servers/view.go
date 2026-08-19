// This file deals with server view implementation
package servers

import tea "charm.land/bubbletea/v2"

// View renders the selected server.
func (m Model) View() tea.View {
	service := m.Engine.Service
	return tea.NewView("Server\n\n" + service.Name + " (" + string(service.Lang) + ")")
}
