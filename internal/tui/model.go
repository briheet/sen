package tui

import (
	"charm.land/bubbles/v2/spinner"
	"github.com/briheet/sen/internal/tui/components"
	"github.com/briheet/sen/internal/tui/components/footer"
	"github.com/briheet/sen/internal/tui/components/header"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/components/notifications"
	"github.com/briheet/sen/internal/tui/context"

	"github.com/briheet/sen/internal/tui/styles"
)

// model contains the engine data rendered by the TUI.
type model struct {
	// program's context (metrics, business stuff)
	ctx *context.ProgramContext

	header header.Model // base header of the application
	footer footer.Model // base footer of the application

	// infobar contains over config loaded configs and themes
	// this keeps user data in front of it, let him see and debug himself
	// this is always hidden component. TODO: Should be opened via shift+:+h
	infobar components.Info

	// workspace wide keys to work with
	keys *keys.Model

	// workspace wide notification.
	// this will help use by telling application crashes, any other issues or such
	notifications notifications.Model

	taskSpinner spinner.Model // use this spinner for tasks loading and such

	// TODO: Please add themes in internal/tui/styles and let people choose at runtime
	activeTheme styles.Theme // active theme for UI

	// dimensions of the model
	width  int
	height int
}
