// Package db contains the shared TUI page for database services.
package db

import (
	"io"

	"charm.land/bubbles/v2/paginator"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/pages/db/metrics"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model owns a database graph and its provider-specific metrics dashboard.
type Model struct {
	Service config.Service
	Engine  *engine.Engine

	dump           io.Writer
	graphs         [2]graph.Model
	metrics        metrics.Model
	pager          paginator.Model
	width          int
	height         int
	viewport       pages.ViewportMsg
	pending        int
	telemetry      uint64
	revision       uint64
	obscured       bool
	showMetrics    bool
	canvas         *lipgloss.Canvas
	view           tea.View
	viewRevision   uint64
	viewValid      bool
	indicator      string
	indicatorPage  int
	indicatorWidth int
}

// New creates a database page from a built engine.
func New(target *engine.Engine, dump io.Writer) *Model {
	return NewWithTheme(target, dump, styles.Zakura)
}

// NewWithTheme creates a database page using theme.
func NewWithTheme(target *engine.Engine, dump io.Writer, theme styles.Theme) *Model {
	statements, tables := databaseViews(target.Graph)
	return &Model{
		Service: target.Service,
		Engine:  target,
		dump:    dump,
		graphs: [2]graph.Model{
			graph.NewWithTheme(target.Service.Name+":statements", graph.FunctionGraph, statements, dump, theme),
			graph.NewWithTheme(target.Service.Name+":tables", graph.FunctionGraph, tables, dump, theme),
		},
		metrics: metrics.NewWithTheme(target.Service.Provider, target.Graph, theme),
		pager:   databasePager(theme),
		pending: -1,
	}
}

// FromService creates a placeholder for a database without a live engine.
func FromService(service config.Service) *Model {
	return FromServiceWithTheme(service, styles.Zakura)
}

// FromServiceWithTheme creates a themed placeholder without a live engine.
func FromServiceWithTheme(service config.Service, theme styles.Theme) *Model {
	return &Model{
		Service: service,
		metrics: metrics.NewWithTheme(service.Provider, nil, theme),
		pager:   databasePager(theme),
		pending: -1,
	}
}

func databasePager(theme styles.Theme) paginator.Model {
	pager := paginator.New(paginator.WithTotalPages(2))
	pager.Type = paginator.Dots
	pager.ActiveDot = lipgloss.NewStyle().Foreground(theme.NodeActive).Render(" ● ")
	pager.InactiveDot = lipgloss.NewStyle().Foreground(theme.Border).Render(" ○ ")
	// Page changes are mouse-driven; arrows remain available to workspace tabs.
	pager.KeyMap.PrevPage.SetEnabled(false)
	pager.KeyMap.NextPage.SetEnabled(false)
	return pager
}

// Name returns the configured service name.
func (m *Model) Name() string { return m.Service.Name }

// Type returns the configured service type.
func (m *Model) Type() config.ServiceType { return m.Service.Type }

// Revision changes when the page's terminal layer changes.
func (m *Model) Revision() uint64 {
	revision := m.revision
	for index := range m.graphs {
		revision += m.graphs[index].Revision()
	}
	return revision
}
