package pages

import "charm.land/lipgloss/v2"

// Overlay reuses a cell canvas while composing a centered panel over content.
func Overlay(canvas **lipgloss.Canvas, width, height int, content, panel string) string {
	if *canvas == nil {
		*canvas = lipgloss.NewCanvas(width, height)
	} else {
		(*canvas).Resize(width, height)
		(*canvas).Clear()
	}
	(*canvas).Compose(lipgloss.NewCompositor(
		lipgloss.NewLayer(content),
		lipgloss.NewLayer(panel).
			X(max(0, (width-lipgloss.Width(panel))/2)).
			Y(max(0, (height-lipgloss.Height(panel))/2)),
	))
	return (*canvas).Render()
}
