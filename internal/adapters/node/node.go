// Package node integrates Node.js source analysis and runtime collection.
package node

import (
	"context"
	"path/filepath"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/node/analysis"
	targetruntime "github.com/briheet/senbon/internal/adapters/node/runtime"
	"github.com/briheet/senbon/internal/model"
)

// Adapter analyzes Node.js applications.
type Adapter struct{}

var _ adapters.Application = (*Adapter)(nil)

// Analyze loads and converts a Node.js application into the normalized graph.
func (*Adapter) Analyze(ctx context.Context, sourcePath string) (*model.StaticGraph, string, error) {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, "", err
	}
	graph, err := analysis.Analyze(ctx, sourcePath)
	if err != nil {
		return nil, "", err
	}
	return graph, sourcePath, nil
}

// Open prepares the instrumented Node.js target.
func (*Adapter) Open(ctx context.Context, sourcePath string) (adapters.Runtime, error) {
	return targetruntime.NewRuntime(ctx, sourcePath)
}
