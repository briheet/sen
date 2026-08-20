// Package rust integrates Cargo applications with sen.
package rust

import (
	"context"
	"errors"
	"sync"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/rust/analysis"
	rustbuild "github.com/briheet/sen/internal/adapters/rust/build"
	targetruntime "github.com/briheet/sen/internal/adapters/rust/runtime"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/model"
)

// Adapter owns the isolated Cargo build shared by analysis and runtime setup.
type Adapter struct {
	mode config.TokioConsoleMode

	mu       sync.Mutex
	prepared *rustbuild.Prepared
	symbols  *analysis.Symbols
}

var _ adapters.Application = (*Adapter)(nil)

// New returns a Rust adapter configured for the selected Tokio Console mode.
func New(mode config.TokioConsoleMode) *Adapter { return &Adapter{mode: mode} }

// Analyze builds the selected Cargo binary and derives its source graph from DWARF.
func (a *Adapter) Analyze(ctx context.Context, sourcePath string, buildArgs []string) (*model.StaticGraph, string, error) {
	prepared, err := rustbuild.Prepare(ctx, sourcePath, buildArgs, a.mode, adapters.Output{})
	if err != nil {
		return nil, "", err
	}
	graph, symbols, err := analysis.Analyze(prepared.Binary, prepared.Workspace, prepared.Package)
	if err != nil {
		_ = prepared.Cleanup()
		return nil, "", err
	}
	a.mu.Lock()
	a.prepared, a.symbols = prepared, symbols
	a.mu.Unlock()
	return graph, prepared.Workspace, nil
}

// Open creates a runtime for the artifact prepared during analysis.
func (a *Adapter) Open(_ context.Context, _ string, _, runArgs []string, output adapters.Output) (adapters.Runtime, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.prepared == nil || a.symbols == nil {
		return nil, errors.New("Rust runtime opened before analysis")
	}
	return targetruntime.New(a.prepared, a.symbols, runArgs, output), nil
}
