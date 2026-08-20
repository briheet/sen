// Package metrics renders Redis telemetry for a key-value service page.
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
	historyLimit  = 120
	historyWindow = 30 * time.Second
)

type sample struct {
	At                time.Time
	Operations        float64
	Connections       float64
	NetworkInput      float64
	NetworkOutput     float64
	CacheHitRatio     float64
	UsedMemory        uint64
	Clients           uint64
	hasConnectionRate bool
	hasInputRate      bool
	hasOutputRate     bool
	hasCacheRatio     bool
}

// Model contains the current Redis snapshot and a bounded rolling history.
type Model struct {
	metrics   model.RedisMetrics
	history   []sample
	sampledAt time.Time
	style     Style
	width     int
	height    int
	offset    int
}

// Style contains the presentation styles used by the Redis dashboard.
type Style struct {
	Root      lipgloss.Style
	Section   lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Activity  lipgloss.Style
	Memory    lipgloss.Style
	Column    lipgloss.Style
	Component lipgloss.Style
	Separator lipgloss.Style
}

// New creates a Redis metrics panel backed by an optional runtime graph.
func New(source *model.RuntimeGraph) Model {
	result := Model{
		history: make([]sample, 0, historyLimit),
		style: Style{
			Root:      styles.Panel(styles.Zakura).Padding(0),
			Section:   lipgloss.NewStyle().Foreground(styles.Zakura.Secondary).Bold(true),
			Label:     lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
			Value:     lipgloss.NewStyle().Foreground(styles.Zakura.Text),
			Activity:  lipgloss.NewStyle().Foreground(styles.Zakura.CPU),
			Memory:    lipgloss.NewStyle().Foreground(styles.Zakura.Memory),
			Column:    lipgloss.NewStyle(),
			Component: lipgloss.NewStyle().Padding(0, 1),
			Separator: lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
		},
	}
	if source != nil {
		result.metrics = source.Snapshot().Metrics.Redis
	}
	return result
}

// ApplySnapshot stores the latest Redis metrics and derives per-second rates.
func (m *Model) ApplySnapshot(metrics model.RuntimeMetrics, sampledAt time.Time) {
	redis := metrics.Redis
	current := sample{
		At:         sampledAt,
		Operations: float64(redis.InstantaneousOps),
		UsedMemory: redis.UsedMemory,
		Clients:    redis.ConnectedClients,
	}
	if !m.sampledAt.IsZero() && sampledAt.After(m.sampledAt) {
		seconds := sampledAt.Sub(m.sampledAt).Seconds()
		previous := m.metrics
		current.Connections, current.hasConnectionRate = counterRate(
			redis.TotalConnectionsReceived, previous.TotalConnectionsReceived, seconds,
		)
		current.NetworkInput, current.hasInputRate = counterRate(
			redis.NetworkInputBytes, previous.NetworkInputBytes, seconds,
		)
		current.NetworkOutput, current.hasOutputRate = counterRate(
			redis.NetworkOutputBytes, previous.NetworkOutputBytes, seconds,
		)
		if redis.KeyspaceHits >= previous.KeyspaceHits && redis.KeyspaceMisses >= previous.KeyspaceMisses {
			hits := redis.KeyspaceHits - previous.KeyspaceHits
			misses := redis.KeyspaceMisses - previous.KeyspaceMisses
			if total := hits + misses; total > 0 {
				current.CacheHitRatio = float64(hits) / float64(total) * 100
				current.hasCacheRatio = true
			}
		}
	}
	m.metrics = redis
	m.sampledAt = sampledAt
	m.appendHistory(current)
}

func counterRate(current, previous uint64, seconds float64) (float64, bool) {
	if seconds <= 0 || current < previous {
		return 0, false
	}
	return float64(current-previous) / seconds, true
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

// Update tracks modal dimensions, terminal background, and scrolling.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(0, msg.Width)
		m.height = max(0, msg.Height)
	case tea.BackgroundColorMsg:
		for _, style := range []*lipgloss.Style{
			&m.style.Root, &m.style.Section, &m.style.Label, &m.style.Value,
			&m.style.Activity, &m.style.Memory, &m.style.Column,
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
	const compactContentHeight = 23
	return max(0, compactContentHeight-max(0, m.height-2))
}

func (m Model) latestSample() sample {
	if len(m.history) == 0 {
		return sample{}
	}
	return m.history[len(m.history)-1]
}
