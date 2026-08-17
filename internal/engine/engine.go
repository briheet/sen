package engine

import (
	"context"

	"github.com/briheet/senbon/internal/analysis"
	"github.com/briheet/senbon/internal/helpers"
	"github.com/briheet/senbon/internal/model"
	targetruntime "github.com/briheet/senbon/internal/runtime"
)

const (
	CwdLimiter = "."
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

	// Load packages and handle errors if so
	pkgs, err := loadPackages(ctx, sourcePath)
	if err != nil {
		return nil, err
	}

	// Build ssa, Analyze via rta and return internal graph representation
	static, err := analysis.GetGraph(pkgs)
	if err != nil {
		return nil, err
	}

	observed, err := targetruntime.NewRuntime(ctx, sourcePath)
	if err != nil {
		return nil, err
	}
	merged := model.BuildRuntimeGraph(pkgs[0].Module.Path, static)

	return &Engine{
		Runtime: observed,
		Graph:   merged,
	}, nil
}

// Snapshot returns a complete runtime update for the TUI to apply.
func (e *Engine) Snapshot() model.RuntimeUpdate {
	return e.Graph.BuildUpdate(e.Runtime)
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
	return nil
}
