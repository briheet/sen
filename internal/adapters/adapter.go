// Package adapters defines capabilities implemented by target integrations.
package adapters

import (
	"context"
	"io"

	"github.com/briheet/sen/internal/model"
)

// Output receives target process output without writing to the terminal.
type Output struct {
	Stdout io.Writer
	Stderr io.Writer
}

const (
	// GoTarget identifies the Go adapter.
	GoTarget = "go"
	// NodeTarget identifies the Node.js adapter.
	NodeTarget = "node"
	// RustTarget identifies the Rust adapter.
	RustTarget = "rust"
	// RedisTarget identifies the Redis adapter.
	RedisTarget = "redis"
	// PostgresTarget identifies the PostgreSQL adapter.
	PostgresTarget = "postgres"
)

// Analyzer builds a normalized static graph and returns its source namespace.
type Analyzer interface {
	Analyze(ctx context.Context, sourcePath string, buildArgs []string) (*model.StaticGraph, string, error)
}

// Application provides analysis and runtime support for one target language.
type Application interface {
	Analyzer
	Open(ctx context.Context, sourcePath string, buildArgs, runArgs []string, output Output) (Runtime, error)
}

// Runtime collects normalized data from a target process.
type Runtime interface {
	Start(context.Context) error
	Collect(context.Context) (model.Observation, error)
	Wait() error
	Stop() error
	Cleanup() error
}
