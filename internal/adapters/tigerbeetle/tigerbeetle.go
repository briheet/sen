// Package tigerbeetle observes a running TigerBeetle cluster through its
// experimental DogStatsD telemetry endpoint.
package tigerbeetle

import (
	"context"
	"strings"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/tigerbeetle/analysis"
	targetruntime "github.com/briheet/sen/internal/adapters/tigerbeetle/runtime"
	"github.com/briheet/sen/internal/model"
)

// Adapter attaches to TigerBeetle; it never owns the database process.
type Adapter struct{}

var _ adapters.Application = (*Adapter)(nil)

// Analyze builds stable operation and replica topology from configuration.
func (*Adapter) Analyze(_ context.Context, source string, _ []string) (*model.StaticGraph, string, error) {
	return analysis.BuildGraph(strings.Split(source, ",")), analysis.ModulePath, nil
}

// Open returns a UDP collector for TigerBeetle's StatsD stream.
func (*Adapter) Open(_ context.Context, address string, _, _ []string, output adapters.Output) (adapters.Runtime, error) {
	return targetruntime.NewCollector(address, output.Stderr), nil
}
