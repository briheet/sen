// Package adapters defines capabilities implemented by target integrations.
package adapters

import (
	"context"
	"time"

	"github.com/briheet/senbon/internal/model"
)

// Analyzer builds a normalized static graph and returns its source namespace.
type Analyzer interface {
	Analyze(context.Context, string) (*model.StaticGraph, string, error)
}

// Runner manages a target process lifecycle.
type Runner interface {
	Run() error
	Stop() error
	Cleanup() error
}

// MetricsCollector collects process-wide runtime metrics.
type MetricsCollector interface {
	CollectMetrics(context.Context) error
}

// Profiler collects a named profile over a duration.
type Profiler interface {
	CollectProfile(context.Context, string, time.Duration) error
}

// Tracer collects runtime events over a duration.
type Tracer interface {
	CollectTrace(context.Context, time.Duration) error
}
