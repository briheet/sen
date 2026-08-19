package tui

import (
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/tui/components/footer"
	"github.com/briheet/sen/internal/tui/components/header"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/context"

	"github.com/briheet/sen/internal/tui/styles"
)

// model contains the engine data rendered by the TUI.
type model struct {
	dump io.Writer // receives Bubble Tea messages when DEBUG is set
	view tea.View  // cached terminal text; graph pixels are uploaded separately

	// program's context (metrics, business stuff)
	ctx *context.ProgramContext

	header header.Model // base header of the application
	footer footer.Model // base footer of the application

	// workspace wide keys to work with
	keys *keys.Model

	// TODO: Please add themes in internal/tui/styles and let people choose at runtime
	activeTheme styles.Theme // active theme for UI

	// dimensions of the model
	width  int
	height int
}
