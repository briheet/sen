// Package factory resolves supported target adapters.
package factory

import (
	"errors"

	"github.com/briheet/senbon/internal/adapters"
	golangadapter "github.com/briheet/senbon/internal/adapters/golang"
	nodeadapter "github.com/briheet/senbon/internal/adapters/node"
	zigadapter "github.com/briheet/senbon/internal/adapters/zig"
)

// ErrUnsupportedTarget reports an unknown language.
var ErrUnsupportedTarget = errors.New("unsupported target")

// Application returns the adapter for a language.
func Application(language string) (adapters.Application, error) {
	switch language {
	case adapters.GoTarget:
		return new(golangadapter.Adapter), nil
	case adapters.NodeTarget:
		return new(nodeadapter.Adapter), nil
	case adapters.ZigTarget:
		return new(zigadapter.Adapter), nil
	default:
		return nil, ErrUnsupportedTarget
	}
}
