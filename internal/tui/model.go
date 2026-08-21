package tui

import (
	"io"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/components/statusbar"
	"github.com/briheet/sen/internal/tui/context"

	"github.com/briheet/sen/internal/tui/styles"
)

// model contains the engine data rendered by the TUI.
type model struct {
	dump io.Writer // receives Bubble Tea messages when DEBUG is set
	view tea.View  // cached terminal text; graph pixels are uploaded separately
	body tea.View

	bodyPage     string
	bodyRevision uint64
	bodyWidth    int
	bodyHeight   int
	bodyEpoch    uint64
	renderEpoch  uint64
	canvas       *lipgloss.Canvas

	// program's context (metrics, business stuff)
	ctx *context.ProgramContext

	statusbar statusbar.Model // workspace navigation and help

	// workspace wide keys to work with
	keys *keys.Model

	activeTheme styles.Theme // active theme for UI

	// dimensions of the model
	width  int
	height int
}
