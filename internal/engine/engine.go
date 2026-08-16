package engine

import (
	"context"

	"github.com/briheet/senbon/internal/graph"
	"github.com/briheet/senbon/internal/helpers"
)

const (
	CwdLimiter = "."
)

// Main engine interface. Couldn't come up with good naming :)
type Driver interface {
	Run() error
}

// This is the main Engine of the program. It verifies, loads, builds,
// performs analysis, setups logs, instrumentation, tui, etc.
// It conforms to Driver interface.
type Engine struct{}

// Have compile time check
var _ Driver = (*Engine)(nil)

// Initilizes a new Engine to work with.
func NewEngine(ctx context.Context, sourcePath string) (*Engine, error) {
	// Validate sourcePath
	if err := helpers.ValidateSourcePath(sourcePath); err != nil {
		return nil, err
	}

	// Load packages and handle errors if so
	pkgs, err := loadPackages(ctx, sourcePath)
	if err != nil {
		return nil, err
	}

	// Build ssa, Analyze via rta and return internal graph representation
	_, err = graph.GetGraph(pkgs)
	if err != nil {
		return nil, err
	}

	return &Engine{}, nil
}

func (e *Engine) Run() error {
	return nil
}
