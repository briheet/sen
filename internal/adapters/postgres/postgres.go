// Package postgres integrates observation of a running PostgreSQL server with
// Senbon over its built-in statistics views.
package postgres

import (
	"context"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/postgres/analysis"
	targetruntime "github.com/briheet/sen/internal/adapters/postgres/runtime"
	"github.com/briheet/sen/internal/model"
)

// Adapter observes a running PostgreSQL server.
type Adapter struct{}

var _ adapters.Application = (*Adapter)(nil)

// Analyze connects to the server and builds a synthetic graph of its current
// statements and tables. The connection string is the "source path".
func (*Adapter) Analyze(ctx context.Context, dsn string, _ []string) (*model.StaticGraph, string, error) {
	return analysis.Analyze(ctx, dsn)
}

// Open returns a collector dialing the running server at the given connection
// string.
func (*Adapter) Open(_ context.Context, dsn string, _, _ []string, _ adapters.Output) (adapters.Runtime, error) {
	return targetruntime.NewCollector(dsn), nil
}
