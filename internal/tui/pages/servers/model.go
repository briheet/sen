package servers

import (
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
)

// Model contains the engine state for one server page.
type Model struct {
	Engine *engine.Engine
}

// New creates a server model from a built engine.
func New(target *engine.Engine) Model {
	return Model{Engine: target}
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Engine.Service.Name }

// Type returns the configured service type.
func (m Model) Type() config.ServiceType { return m.Engine.Service.Type }
