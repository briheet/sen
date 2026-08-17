// Package engine joins a target adapter with sen's normalized model.
package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/factory"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/helpers"
	"github.com/briheet/sen/internal/model"
)

const collectionTimeout = time.Minute

// Engine owns one configured service, its runtime, and its graph.
type Engine struct {
	Service config.Service
	Runtime adapters.Runtime
	Graph   *model.RuntimeGraph

	mu       sync.RWMutex
	revision atomic.Uint64
}

// NewEngine resolves, analyzes, and opens a configured service.
func NewEngine(ctx context.Context, service config.Service, output adapters.Output) (*Engine, error) {
	target := string(service.Lang)
	source := service.Path
	if service.Type == config.ServiceTypeKV {
		target = string(service.Provider)
		source = service.Address
	} else {
		if target != adapters.PostgresTarget {
			if err := helpers.ValidateSourcePath(source); err != nil {
				return nil, err
			}
		}
	}
	application, err := factory.Application(target)
	if err != nil {
		return nil, err
	}
	static, namespace, err := application.Analyze(ctx, source, service.BuildArgs)
	if err != nil {
		return nil, err
	}
	runtime, err := application.Open(ctx, source, service.BuildArgs, service.RunArgs, output)
	if err != nil {
		return nil, err
	}
	return &Engine{Service: service, Runtime: runtime, Graph: model.BuildRuntimeGraph(namespace, static)}, nil
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
	update := e.Graph.BuildUpdate(observation.Metrics, observation.Profiles, observation.Trace)
	e.mu.Lock()
	e.Graph.ApplyUpdate(update)
	e.mu.Unlock()
	e.revision.Add(1)
	return nil
}

// Revision changes after each complete metrics, profile, and trace window.
func (e *Engine) Revision() uint64 { return e.revision.Load() }

// Snapshot returns the latest completed runtime window.
func (e *Engine) Snapshot() model.RuntimeSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Graph.Snapshot()
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

// Run continuously collects runtime windows until the target exits.
func (e *Engine) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := e.Start(ctx); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() { done <- e.Wait() }()
	for {
		collected := make(chan error, 1)
		go func() {
			window, stop := context.WithTimeout(ctx, collectionTimeout)
			defer stop()
			collected <- e.Refresh(window)
		}()

		select {
		case err := <-done:
			cancel()
			<-collected
			return err
		case err := <-collected:
			if err == nil {
				continue
			}
			select {
			case exitErr := <-done:
				return exitErr
			default:
			}
			_ = e.Stop()
			<-done
			return err
		}
	}
}
