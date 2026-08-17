package engine

import (
	"context"
	"log"
	"time"

	"github.com/briheet/senbon/internal/adapters/golang"
	targetruntime "github.com/briheet/senbon/internal/adapters/golang/runtime"
	"github.com/briheet/senbon/internal/helpers"
	"github.com/briheet/senbon/internal/model"
)

// Main engine interface. Couldn't come up with good naming :)
type Driver interface {
	Run() error
}

// Engine owns the target runtime and the merged graph exposed to the TUI.
type Engine struct {
	Runtime *targetruntime.Runtime
	Graph   *model.RuntimeGraph
}

// Have compile time check
var _ Driver = (*Engine)(nil)

// Initilizes a new Engine to work with.
func NewEngine(ctx context.Context, sourcePath string) (*Engine, error) {
	// Validate sourcePath
	if err := helpers.ValidateSourcePath(sourcePath); err != nil {
		return nil, err
	}

	static, namespace, err := new(golang.Adapter).Analyze(ctx, sourcePath)
	if err != nil {
		return nil, err
	}

	observed, err := targetruntime.NewRuntime(ctx, sourcePath)
	if err != nil {
		return nil, err
	}
	merged := model.BuildRuntimeGraph(namespace, static)

	return &Engine{
		Runtime: observed,
		Graph:   merged,
	}, nil
}

// Snapshot returns a complete runtime update for the TUI to apply.
func (e *Engine) Snapshot() model.RuntimeUpdate {
	return e.Graph.BuildUpdate(e.Runtime.Metrics, e.Runtime.Profiles, e.Runtime.Trace)
}

// MetricsUpdate returns the latest process metrics.
func (e *Engine) MetricsUpdate() model.RuntimeUpdate {
	return e.Graph.BuildMetricsUpdate(e.Runtime.Metrics)
}

// ProfileUpdate returns one named profile, replacing its previous data.
func (e *Engine) ProfileUpdate(name string) model.RuntimeUpdate {
	return e.Graph.BuildProfileUpdate(name, e.Runtime.Profiles[name])
}

// TraceUpdate returns the latest trace, replacing its previous data.
func (e *Engine) TraceUpdate() model.RuntimeUpdate {
	return e.Graph.BuildTraceUpdate(e.Runtime.Trace)
}

func (e *Engine) Run() error {
	done := make(chan error, 1)
	go func() { done <- e.Runtime.Process.Run() }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for {
		if err := e.Runtime.CollectMetrics(ctx); err == nil {
			break
		}
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			_ = e.Runtime.Process.Stop()
			<-done
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	log.Printf("senbon: collector ready")

	started := time.Now()
	results := make(chan error, 2)
	go func() { results <- e.Runtime.CollectProfile(ctx, "cpu", time.Second) }()
	go func() { results <- e.Runtime.CollectTrace(ctx, time.Second) }()
	for range 2 {
		if err := <-results; err != nil {
			_ = e.Runtime.Process.Stop()
			<-done
			return err
		}
	}
	log.Printf("senbon: cpu profile and trace collected in %s", time.Since(started))

	if err := e.Runtime.CollectMetrics(ctx); err != nil {
		_ = e.Runtime.Process.Stop()
		<-done
		return err
	}

	e.Graph.ApplyUpdate(e.Snapshot())
	mapped := 0
	for _, node := range e.Graph.Nodes {
		if len(node.Metrics) != 0 {
			mapped++
		}
	}
	log.Printf("senbon: heap=%d objects=%d cpu_samples=%d trace_events=%d mapped_nodes=%d",
		e.Graph.Global.Process.LiveHeap,
		e.Graph.Global.Process.CurrHeapObjects,
		len(e.Runtime.Profiles["cpu"].Samples),
		len(e.Runtime.Trace.Events),
		mapped,
	)
	return <-done
}
