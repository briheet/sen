package servers

import (
	"io"

	"charm.land/bubbles/v2/paginator"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/pages/servers/metrics"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model contains the engine state for one server page.
type Model struct {
	Engine *engine.Engine

	graphs      [2]graph.Model
	metrics     metrics.Model
	pager       paginator.Model
	width       int
	height      int
	viewport    pages.ViewportMsg
	pending     int
	telemetry   uint64
	revision    uint64
	obscured    bool
	showMetrics bool
}

// New creates a server model from a built engine.
func New(target *engine.Engine, dump io.Writer) Model {
	pager := paginator.New(paginator.WithTotalPages(2))
	pager.Type = paginator.Dots
	pager.ActiveDot = lipgloss.NewStyle().Foreground(styles.Zakura.NodeActive).Render(" ● ")
	pager.InactiveDot = lipgloss.NewStyle().Foreground(styles.Zakura.Border).Render(" ○ ")
	// Page changes are mouse-driven; arrows remain available to workspace tabs.
	pager.KeyMap.PrevPage.SetEnabled(false)
	pager.KeyMap.NextPage.SetEnabled(false)
	return Model{
		Engine: target,
		graphs: [2]graph.Model{
			graph.New(target.Service.Name+":functions", graph.FunctionGraph, target.Graph, dump),
			graph.New(target.Service.Name+":files", graph.FileGraph, target.Graph, dump),
		},
		metrics: metrics.New(target.Graph, target.Service.Lang),
		pager:   pager,
		pending: -1,
	}
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Engine.Service.Name }

// Type returns the configured service type.
func (m Model) Type() config.ServiceType { return m.Engine.Service.Type }

// Revision changes when the server page's terminal text changes.
func (m Model) Revision() uint64 {
	revision := m.revision
	for index := range m.graphs {
		revision += m.graphs[index].Revision()
	}
	return revision
}
