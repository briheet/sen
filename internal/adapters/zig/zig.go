// Package zig integrates Zig source analysis and runtime collection.
package zig

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/zig/analysis"
	"github.com/briheet/senbon/internal/adapters/zig/runtime"
	"github.com/briheet/senbon/internal/model"
)

// Adapter analyzes Zig applications.
type Adapter struct {
	project *analysis.Project
}

var _ adapters.Application = (*Adapter)(nil)

// Analyze loads and converts a Zig application into the normalized graph.
func (a *Adapter) Analyze(ctx context.Context, sourcePath string) (*model.StaticGraph, string, error) {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, "", err
	}
	project, err := analysis.Analyze(ctx, sourcePath)
	if err != nil {
		return nil, "", err
	}
	a.project = project
	return project.Graph, sourcePath, nil
}

// Open builds the instrumented Zig target.
func (a *Adapter) Open(ctx context.Context, sourcePath string) (adapters.Runtime, error) {
	if a.project == nil {
		return nil, errors.New("zig: analyze the project before opening")
	}
	return runtime.NewRuntime(ctx, sourcePath, a.project)
}
