// Package kv contains TUI state for key-value services.
package kv

import (
	"io"

	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/components/graph"
	redismetrics "github.com/briheet/sen/internal/tui/pages/kv/metrics"
)

// Model contains the command graph and Redis telemetry for one KV service.
type Model struct {
	Service config.Service
	Engine  *engine.Engine

	dump        io.Writer
	graph       graph.Model
	metrics     redismetrics.Model
	width       int
	height      int
	telemetry   uint64
	revision    uint64
	obscured    bool
	showMetrics bool
}

// New creates a key-value model from a built engine.
func New(target *engine.Engine, dump io.Writer) Model {
	return Model{
		Service: target.Service,
		Engine:  target,
		dump:    dump,
		graph:   graph.New(target.Service.Name+":commands", graph.FunctionGraph, target.Graph, dump),
		metrics: redismetrics.New(target.Graph),
	}
}

// FromService creates a temporary model for integrations without engines.
func FromService(service config.Service) Model {
	return Model{Service: service}
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Service.Name }

// Type returns the configured service type.
func (m Model) Type() config.ServiceType { return m.Service.Type }

// Revision changes when the page's native terminal layer changes.
func (m Model) Revision() uint64 { return m.revision + m.graph.Revision() }
