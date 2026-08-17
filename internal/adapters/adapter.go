// Package adapters defines capabilities implemented by target integrations.
package adapters

import (
	"context"

	"github.com/briheet/senbon/internal/model"
)

const (
	// GoTarget identifies the Go adapter.
	GoTarget = "go"
	// NodeTarget identifies the Node.js adapter.
	NodeTarget = "node"
	// OcamlTarget identifies the OCaml adapter.
	OcamlTarget = "ocaml"
)

// Analyzer builds a normalized static graph and returns its source namespace.
type Analyzer interface {
	Analyze(context.Context, string) (*model.StaticGraph, string, error)
}

// Application provides analysis and runtime support for one target language.
type Application interface {
	Analyzer
	Open(context.Context, string) (Runtime, error)
}

// Runtime collects normalized data from a target process.
type Runtime interface {
	Start(context.Context) error
	Collect(context.Context) (model.Observation, error)
	Wait() error
	Stop() error
	Cleanup() error
}
