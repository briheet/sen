package kv

import tea "charm.land/bubbletea/v2"

// View renders the selected key-value service.
func (m Model) View() tea.View {
	return tea.NewView("Key-value store\n\n" + m.Service.Name + " (" + string(m.Service.Provider) + ")\n" + m.Service.Address)
}
