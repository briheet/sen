// Package tigerbeetle contains the provider-specific TigerBeetle TUI page.
package tigerbeetle

import (
	"io"

	"charm.land/bubbles/v2/paginator"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/pages/tigerbeetle/metrics"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model owns TigerBeetle's operation and replica graphs.
type Model struct {
	Service config.Service
	Engine  *engine.Engine

	dump        io.Writer
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

// New creates a TigerBeetle page from a built engine.
func New(target *engine.Engine, dump io.Writer) Model {
	operations, replicas := tigerBeetleViews(target.Graph)
	return Model{
		Service: target.Service, Engine: target, dump: dump,
		graphs: [2]graph.Model{
			graph.New(target.Service.Name+":operations", graph.FunctionGraph, operations, dump),
			graph.New(target.Service.Name+":replicas", graph.FunctionGraph, replicas, dump),
		},
		metrics: metrics.New(target.Graph, len(target.Service.Addresses)), pager: pagePager(), pending: -1,
	}
}

// FromService creates a placeholder when engine construction failed.
func FromService(service config.Service) Model {
	return Model{Service: service, metrics: metrics.New(nil, len(service.Addresses)), pager: pagePager(), pending: -1}
}

func pagePager() paginator.Model {
	pager := paginator.New(paginator.WithTotalPages(2))
	pager.Type = paginator.Dots
	pager.ActiveDot = lipgloss.NewStyle().Foreground(styles.Zakura.NodeActive).Render(" ● ")
	pager.InactiveDot = lipgloss.NewStyle().Foreground(styles.Zakura.Border).Render(" ○ ")
	pager.KeyMap.PrevPage.SetEnabled(false)
	pager.KeyMap.NextPage.SetEnabled(false)
	return pager
}

// Name returns the configured service name.
func (m Model) Name() string { return m.Service.Name }

// Type returns the database service type.
func (m Model) Type() config.ServiceType { return m.Service.Type }

// Revision changes when the terminal or native graph layer changes.
func (m Model) Revision() uint64 {
	revision := m.revision
	for index := range m.graphs {
		revision += m.graphs[index].Revision()
	}
	return revision
}
