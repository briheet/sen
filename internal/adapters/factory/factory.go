// Package factory resolves supported target adapters.
package factory

import (
	"errors"

	"github.com/briheet/sen/internal/adapters"
	golangadapter "github.com/briheet/sen/internal/adapters/golang"
	nodeadapter "github.com/briheet/sen/internal/adapters/node"
	postgresadapter "github.com/briheet/sen/internal/adapters/postgres"
	redisadapter "github.com/briheet/sen/internal/adapters/redis"
	rustadapter "github.com/briheet/sen/internal/adapters/rust"
	"github.com/briheet/sen/internal/config"
)

// ErrUnsupportedTarget reports an unknown language.
var ErrUnsupportedTarget = errors.New("unsupported target")

// Application returns the adapter for a language.
func Application(language string, service config.Service) (adapters.Application, error) {
	switch language {
	case adapters.GoTarget:
		return new(golangadapter.Adapter), nil
	case adapters.NodeTarget:
		return new(nodeadapter.Adapter), nil
	case adapters.RustTarget:
		return rustadapter.New(service.TokioConsole), nil
	case adapters.RedisTarget:
		return new(redisadapter.Adapter), nil
	case adapters.PostgresTarget:
		return new(postgresadapter.Adapter), nil
	default:
		return nil, ErrUnsupportedTarget
	}
}
