package servers

import (
	"io"

	"charm.land/bubbles/v2/paginator"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model contains the engine state for one server page.
type Model struct {
	Engine *engine.Engine

	graph  graph.Model
	pager  paginator.Model
	width  int
	height int
}

// New creates a server model from a built engine.
func New(target *engine.Engine, dump io.Writer) Model {
	pager := paginator.New(paginator.WithTotalPages(3))
	pager.Type = paginator.Dots
	pager.ActiveDot = lipgloss.NewStyle().Foreground(styles.Zakura.NodeActive).Render(" ● ")
	pager.InactiveDot = lipgloss.NewStyle().Foreground(styles.Zakura.Border).Render(" ○ ")
	// Page changes are mouse-driven; arrows remain available to workspace tabs.
	pager.KeyMap.PrevPage.SetEnabled(false)
	pager.KeyMap.NextPage.SetEnabled(false)
	return Model{
		Engine: target,
		graph:  graph.New(target.Service.Name, target.Graph, dump),
		pager:  pager,
	}
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Engine.Service.Name }

// Type returns the configured service type.
func (m Model) Type() config.ServiceType { return m.Engine.Service.Type }
