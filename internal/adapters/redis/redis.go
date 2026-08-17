// Package redis integrates observation of a running Redis server with Senbon.
package redis

import (
	"context"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/redis/analysis"
	targetruntime "github.com/briheet/sen/internal/adapters/redis/runtime"
	"github.com/briheet/sen/internal/model"
)

// Adapter observes a running Redis server.
type Adapter struct{}

var _ adapters.Application = (*Adapter)(nil)

// Analyze returns the synthetic Redis command graph.
func (*Adapter) Analyze(ctx context.Context, source string, buildArgs []string) (*model.StaticGraph, string, error) {
	return analysis.BuildGraph(), analysis.ModulePath, nil
}

// Open dials the running Redis server at the given address.
func (*Adapter) Open(ctx context.Context, source string, buildArgs, runArgs []string, output adapters.Output) (adapters.Runtime, error) {
	return targetruntime.NewCollector(source), nil
}
