// Package factory resolves supported target adapters.
package factory

import (
	"errors"

	"github.com/briheet/sen/internal/adapters"
	golangadapter "github.com/briheet/sen/internal/adapters/golang"
	nodeadapter "github.com/briheet/sen/internal/adapters/node"
	postgresadapter "github.com/briheet/sen/internal/adapters/postgres"
	redisadapter "github.com/briheet/sen/internal/adapters/redis"
	tigerbeetleadapter "github.com/briheet/sen/internal/adapters/tigerbeetle"
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
	case adapters.RedisTarget:
		return new(redisadapter.Adapter), nil
	case adapters.PostgresTarget:
		return new(postgresadapter.Adapter), nil
	case adapters.TigerBeetleTarget:
		return new(tigerbeetleadapter.Adapter), nil
	default:
		return nil, ErrUnsupportedTarget
	}
}
