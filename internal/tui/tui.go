// Package tui presents built sen engines in the terminal.
package tui

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/build"
	"github.com/briheet/sen/internal/config"
)

// Tui owns the terminal model and engines for one run.
type Tui struct {
	model model
	group *build.Group
}

// NewTui builds configured engines and initializes the terminal model.
func NewTui(ctx context.Context, configuration *config.Config) (*Tui, error) {
	group, err := build.New(ctx, configuration)
	if err != nil {
		return nil, err
	}
	return &Tui{
		model: initialModel(group.Engines, configuration.Services, group.LogPath(), group.DebugWriter()),
		group: group,
	}, nil
}

// Run starts the engines and owns their shutdown and cleanup.
func (t *Tui) Run(ctx context.Context) error {
	if path := t.group.DebugPath(); path != "" {
		_, _ = fmt.Fprintln(os.Stderr, "TUI debug log:", path)
	}
	program := tea.NewProgram(t.model, tea.WithContext(ctx))
	done := make(chan error, 1)
	// Engine execution blocks, so keep it outside Bubble Tea's event loop
	go func() {
		err := t.group.Run()
		done <- err
		program.Send(enginesDoneMsg{})
	}()

	_, programErr := program.Run()
	select {
	case runErr := <-done:
		// Natural process completion preserves engine errors for the CLI
		return errors.Join(programErr, runErr, t.group.Cleanup())
	default:
		// A user exit stops processes before their resources and logs are closed
		stopErr := t.group.Stop()
		<-done
		return errors.Join(programErr, stopErr, t.group.Cleanup())
	}
}
