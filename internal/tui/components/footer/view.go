package footer

// View renders short or full workspace help.
func (m Model) View() string {
	return m.help.View(m.keys)
}
