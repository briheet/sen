package servers

import (
	"io"

	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/components/pager"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model contains the engine state for one server page.
type Model struct {
	Engine *engine.Engine

	graph  graph.Model
	pager  pager.Model
	width  int
	height int
}

// New creates a server model from a built engine.
func New(target *engine.Engine, dump io.Writer) Model {
	return Model{
		Engine: target,
		graph:  graph.New(target.Service.Name, target.Graph, dump),
		pager:  pager.New(3, styles.Zakura),
	}
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Engine.Service.Name }

// Type returns the configured service type.
func (m Model) Type() config.ServiceType { return m.Engine.Service.Type }
