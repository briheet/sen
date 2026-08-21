// Package tui presents built sen engines in the terminal.
package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/build"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/styles"
)

// Tui owns the terminal model and engines for one run.
type Tui struct {
	model *model
	group *build.Group
}

// NewTui builds configured engines and initializes the terminal model.
func NewTui(ctx context.Context, configuration *config.Config) (*Tui, error) {
	theme, err := styles.ResolveTheme(configuration.Project.Theme)
	if err != nil {
		return nil, err
	}
	group, err := build.New(ctx, configuration)
	if err != nil {
		return nil, err
	}
	return &Tui{
		model: initialModel(group.Engines, configuration.Services, group.LogPath(), group.DebugWriter(), theme),
		group: group,
	}, nil
}

// Run starts the engines and owns their shutdown and cleanup.
func (t *Tui) Run(ctx context.Context) error {
	if path := t.group.DebugPath(); path != "" {
		_, _ = fmt.Fprintln(os.Stderr, "TUI debug log:", path)
	}
	program := tea.NewProgram(t.model, tea.WithContext(ctx))
	telemetryContext, stopTelemetry := context.WithCancel(ctx)
	var telemetry sync.WaitGroup
	for _, target := range t.group.Engines {
		telemetry.Add(1)
		go forwardTelemetry(telemetryContext, &telemetry, program, target)
	}
	done := make(chan error, 1)
	// Engine execution blocks, so keep it outside Bubble Tea's event loop
	go func() {
		err := t.group.Run()
		done <- err
		program.Send(enginesDoneMsg{})
	}()

	_, programErr := program.Run()
	stopTelemetry()
	telemetry.Wait()
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

func forwardTelemetry(ctx context.Context, wait *sync.WaitGroup, program *tea.Program, target *engine.Engine) {
	defer wait.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-target.Updates():
			program.Send(pages.TelemetryTickMsg{At: time.Now(), Service: target.Service.Name})
		}
	}
}
