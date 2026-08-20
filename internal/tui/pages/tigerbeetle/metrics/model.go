// Package metrics renders TigerBeetle cluster telemetry.
package metrics

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/styles"
)

const (
	screenPercent = 80
	minimumWidth  = 44
	minimumHeight = 10
	historyLimit  = 64
	historyWindow = 5 * time.Minute
	staleAfter    = 25 * time.Second
)

type sample struct {
	At       time.Time
	Requests float64
	Latency  float64
	CacheHit float64
	Queue    float64
}

// Model contains the latest cluster snapshot and bounded five-minute history.
type Model struct {
	metrics   model.TigerBeetleMetrics
	expected  int
	history   []sample
	sampledAt time.Time
	style     Style
	width     int
	height    int
	offset    int
}

// Style contains the dashboard presentation styles.
type Style struct {
	Root      lipgloss.Style
	Section   lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Activity  lipgloss.Style
	Storage   lipgloss.Style
	Healthy   lipgloss.Style
	Warning   lipgloss.Style
	Column    lipgloss.Style
	Component lipgloss.Style
	Separator lipgloss.Style
}

// New creates a metrics panel for the configured replica count.
func New(source *model.RuntimeGraph, expected int) Model {
	result := Model{
		expected: expected, history: make([]sample, 0, historyLimit),
		style: Style{
			Root:     styles.Panel(styles.Zakura).Padding(0),
			Section:  lipgloss.NewStyle().Foreground(styles.Zakura.Secondary).Bold(true),
			Label:    lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
			Value:    lipgloss.NewStyle().Foreground(styles.Zakura.Text),
			Activity: lipgloss.NewStyle().Foreground(styles.Zakura.CPU),
			Storage:  lipgloss.NewStyle().Foreground(styles.Zakura.Memory),
			Healthy:  lipgloss.NewStyle().Foreground(styles.Zakura.Success),
			Warning:  lipgloss.NewStyle().Foreground(styles.Zakura.Warning),
			Column:   lipgloss.NewStyle(), Component: lipgloss.NewStyle().Padding(0, 1),
			Separator: lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
		},
	}
	if source != nil {
		result.metrics = source.Snapshot().Metrics.TigerBeetle
	}
	return result
}

// ApplySnapshot stores a completed native telemetry window.
func (m *Model) ApplySnapshot(metrics model.RuntimeMetrics, sampledAt time.Time) {
	m.metrics = metrics.TigerBeetle
	m.sampledAt = sampledAt
	current := sample{At: sampledAt}
	seconds := m.metrics.Window.Seconds()
	var requests, latency uint64
	for _, operation := range m.metrics.Operations {
		requests += operation.Requests
		latency += uint64(operation.LatencySum)
	}
	if seconds > 0 {
		current.Requests = float64(requests) / seconds
	}
	if requests > 0 {
		current.Latency = float64(latency) / float64(requests) / float64(time.Microsecond)
	}
	var hits, misses uint64
	for _, replica := range m.metrics.Replicas {
		hits += replica.GridCacheHits
		misses += replica.GridCacheMisses
		current.Queue = max(current.Queue, float64(replica.PipelineQueueLength))
	}
	if total := hits + misses; total > 0 {
		current.CacheHit = float64(hits) / float64(total) * 100
	}
	m.appendHistory(current)
}

func (m *Model) appendHistory(current sample) {
	cutoff := current.At.Add(-historyWindow)
	first := 0
	for first < len(m.history) && m.history[first].At.Before(cutoff) {
		first++
	}
	if first > 0 {
		copy(m.history, m.history[first:])
		m.history = m.history[:len(m.history)-first]
	}
	if len(m.history) == historyLimit {
		copy(m.history, m.history[1:])
		m.history = m.history[:historyLimit-1]
	}
	m.history = append(m.history, current)
}

// Update tracks modal dimensions, background color, and compact scrolling.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
	case tea.BackgroundColorMsg:
		for _, style := range []*lipgloss.Style{
			&m.style.Root, &m.style.Section, &m.style.Label, &m.style.Value, &m.style.Activity,
			&m.style.Storage, &m.style.Healthy, &m.style.Warning, &m.style.Column,
			&m.style.Component, &m.style.Separator,
		} {
			*style = style.Background(msg.Color)
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			m.offset = max(0, m.offset-1)
		case "down", "j":
			m.offset = min(m.maxOffset(), m.offset+1)
		}
	}
	m.offset = min(m.offset, m.maxOffset())
	return m, nil
}

// Size returns a centered modal occupying 80 percent of the viewport.
func Size(width, height int) (int, int) {
	return min(width, max(minimumWidth, (width*screenPercent+50)/100)),
		min(height, max(minimumHeight, (height*screenPercent+50)/100))
}

func (m Model) maxOffset() int {
	left, right := m.compactColumns()
	return max(0, max(len(left), len(right))-max(0, m.height-2))
}

func (m Model) stale(replica model.TigerBeetleReplicaMetrics) bool {
	return replica.ObservedAt.IsZero() || m.sampledAt.Sub(replica.ObservedAt) > staleAfter
}
