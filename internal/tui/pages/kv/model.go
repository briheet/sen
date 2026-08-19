// Package kv contains TUI state for key-value services.
package kv

import (
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
)

// Model contains one key-value service configuration.
type Model struct {
	Service config.Service
	Engine  *engine.Engine
}

// New creates a key-value model from a built engine.
func New(target *engine.Engine) Model {
	return Model{Service: target.Service, Engine: target}
}

// FromService creates a temporary model for integrations without engines.
func FromService(service config.Service) Model {
	return Model{Service: service}
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Service.Name }

// Type returns the configured service type.
func (m Model) Type() config.ServiceType { return m.Service.Type }

// Revision changes when the page's rendered text changes.
func (Model) Revision() uint64 { return 0 }
