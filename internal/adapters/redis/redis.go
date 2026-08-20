// Package redis adapts a running Redis server to sen's analysis/runtime API.
package redis

import (
	"context"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/redis/analysis"
	targetruntime "github.com/briheet/sen/internal/adapters/redis/runtime"
	"github.com/briheet/sen/internal/model"
)

// Adapter provides a synthetic command graph and a read-only live collector.
type Adapter struct{}

var _ adapters.Application = (*Adapter)(nil)

// Analyze returns the synthetic Redis command graph.
func (*Adapter) Analyze(context.Context, string, []string) (*model.StaticGraph, string, error) {
	return analysis.BuildGraph(), analysis.ModulePath, nil
}

// Open dials the running Redis server at the given address.
func (*Adapter) Open(_ context.Context, address string, _, _ []string, _ adapters.Output) (adapters.Runtime, error) {
	return targetruntime.NewCollector(address), nil
}
