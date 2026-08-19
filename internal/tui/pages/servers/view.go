// This file deals with server view implementation
package servers

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
)

// View renders the selected server.
func (m Model) View() tea.View {
	contentHeight := max(0, m.height-1)
	content := m.graph.View()
	if m.Engine.Graph != nil {
		switch m.pager.Page {
		case 1:
			metrics := m.Engine.Graph.Global.Process
			content = fmt.Sprintf("Runtime\n\nCPU %.2f   Heap %d B   Goroutines %d",
				metrics.UserCPU, metrics.LiveHeap, metrics.TotalLiveGoroutines)
		case 2:
			trace := m.Engine.Graph.Global.Trace
			content = fmt.Sprintf("Trace\n\nDuration %s   Goroutines %d   Threads %d",
				trace.Duration, trace.Goroutines, trace.Threads)
		}
	}
	if m.pager.Page != 0 {
		content = lipgloss.NewStyle().
			Width(m.width).
			Height(contentHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(styles.Zakura.Text).
			Render(content)
	}
	indicator := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.pager.View())
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, indicator))
	view.MouseMode = tea.MouseModeCellMotion
	return view
}
