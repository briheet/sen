// Package engine joins a target adapter with Senbon's normalized model.
package engine

import (
	"context"
	"time"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/factory"
	"github.com/briheet/senbon/internal/helpers"
	"github.com/briheet/senbon/internal/model"
)

const collectionTimeout = time.Minute

// Engine owns a target runtime and the graph exposed to the TUI.
type Engine struct {
	Runtime adapters.Runtime
	Graph   *model.RuntimeGraph
}

// NewEngine resolves, analyzes, and opens a target application.
func NewEngine(ctx context.Context, sourcePath, language string) (*Engine, error) {
	if err := helpers.ValidateSourcePath(sourcePath); err != nil {
		return nil, err
	}
	application, err := factory.Application(language)
	if err != nil {
		return nil, err
	}
	static, namespace, err := application.Analyze(ctx, sourcePath)
	if err != nil {
		return nil, err
	}
	target, err := application.Open(ctx, sourcePath)
	if err != nil {
		return nil, err
	}
	return &Engine{Runtime: target, Graph: model.BuildRuntimeGraph(namespace, static)}, nil
}

// Start launches the target application.
func (e *Engine) Start(ctx context.Context) error {
	return e.Runtime.Start(ctx)
}

// Refresh collects and applies one runtime snapshot.
func (e *Engine) Refresh(ctx context.Context) error {
	observation, err := e.Runtime.Collect(ctx)
	if err != nil {
		return err
	}
	e.Graph.ApplyUpdate(e.Graph.BuildUpdate(observation.Metrics, observation.Profiles, observation.Trace))
	return nil
}

// Wait blocks until the target exits.
func (e *Engine) Wait() error {
	return e.Runtime.Wait()
}

// Stop terminates the target application.
func (e *Engine) Stop() error {
	return e.Runtime.Stop()
}

// Cleanup removes adapter-owned temporary files.
func (e *Engine) Cleanup() error {
	return e.Runtime.Cleanup()
}

// Run starts the target, collects one snapshot, and waits for exit.
func (e *Engine) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), collectionTimeout)
	defer cancel()
	if err := e.Start(ctx); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() { done <- e.Wait() }()
	collected := make(chan error, 1)
	go func() { collected <- e.Refresh(ctx) }()

	select {
	case err := <-done:
		return err
	case err := <-collected:
		if err != nil {
			_ = e.Stop()
			<-done
			return err
		}
		return <-done
	}
}
