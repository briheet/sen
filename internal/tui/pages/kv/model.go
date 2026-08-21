// Package kv contains TUI state for key-value services.
package kv

import (
	"io"

	"charm.land/bubbles/v2/paginator"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	redismetrics "github.com/briheet/sen/internal/tui/pages/kv/metrics"
	"github.com/briheet/sen/internal/tui/pages/servers/graph"
	"github.com/briheet/sen/internal/tui/styles"
)

// Model contains the command graph and Redis telemetry for one KV service.
type Model struct {
	Service config.Service
	Engine  *engine.Engine

	dump           io.Writer
	graph          graph.Model
	metrics        redismetrics.Model
	pager          paginator.Model
	width          int
	height         int
	telemetry      uint64
	revision       uint64
	obscured       bool
	showMetrics    bool
	canvas         *lipgloss.Canvas
	view           tea.View
	viewRevision   uint64
	viewValid      bool
	indicator      string
	indicatorWidth int
}

// New creates a key-value model from a built engine.
func New(target *engine.Engine, dump io.Writer) *Model {
	return NewWithTheme(target, dump, styles.Zakura)
}

// NewWithTheme creates a key-value model using theme.
func NewWithTheme(target *engine.Engine, dump io.Writer, theme styles.Theme) *Model {
	return &Model{
		Service: target.Service,
		Engine:  target,
		dump:    dump,
		graph:   graph.NewWithTheme(target.Service.Name+":commands", graph.FunctionGraph, target.Graph, dump, theme),
		metrics: redismetrics.NewWithTheme(target.Graph, theme),
		pager:   keyValuePager(theme),
	}
}

// FromService creates a temporary model for integrations without engines.
func FromService(service config.Service) *Model {
	return FromServiceWithTheme(service, styles.Zakura)
}

// FromServiceWithTheme creates a themed placeholder without a live engine.
func FromServiceWithTheme(service config.Service, theme styles.Theme) *Model {
	return &Model{Service: service, metrics: redismetrics.NewWithTheme(nil, theme), pager: keyValuePager(theme)}
}

func keyValuePager(theme styles.Theme) paginator.Model {
	pager := paginator.New(paginator.WithTotalPages(1))
	pager.Type = paginator.Dots
	pager.ActiveDot = lipgloss.NewStyle().Foreground(theme.NodeActive).Render(" ● ")
	pager.InactiveDot = lipgloss.NewStyle().Foreground(theme.Border).Render(" ○ ")
	pager.KeyMap.PrevPage.SetEnabled(false)
	pager.KeyMap.NextPage.SetEnabled(false)
	return pager
}

// Name returns the configured service name.
func (m *Model) Name() string { return m.Service.Name }

// Type returns the configured service type.
func (m *Model) Type() config.ServiceType { return m.Service.Type }

// Revision changes when the page's native terminal layer changes.
func (m *Model) Revision() uint64 { return m.revision + m.graph.Revision() }
