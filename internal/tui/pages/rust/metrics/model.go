// Package metrics renders native Rust and optional Tokio Console measurements.
package metrics

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/styles"
)

const (
	ScreenPercent = 80
	MinimumWidth  = 44
	MinimumHeight = 10
	historyLimit  = 120
	historyWindow = 30 * time.Second
)

type sample struct {
	at        time.Time
	cpu       float64
	rss       uint64
	liveTasks uint64
	hasCPU    bool
	hasRSS    bool
	hasTokio  bool
}

// Model stores the latest Rust measurements and a bounded rolling history.
type Model struct {
	metrics   model.RuntimeMetrics
	history   []sample
	sampledAt time.Time
	style     Style
	width     int
	height    int
	offset    int
}

// Style contains dashboard presentation styles.
type Style struct {
	Root, Section, Label, Value, CPU, Memory, Tokio, Component, Separator lipgloss.Style
}

// New creates a Rust metrics dashboard backed by the runtime graph.
func New(source *model.RuntimeGraph) Model {
	result := Model{
		history: make([]sample, 0, historyLimit),
		style: Style{
			Root:      styles.Panel(styles.Zakura).Padding(0),
			Section:   lipgloss.NewStyle().Foreground(styles.Zakura.Secondary).Bold(true),
			Label:     lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
			Value:     lipgloss.NewStyle().Foreground(styles.Zakura.Text),
			CPU:       lipgloss.NewStyle().Foreground(styles.Zakura.CPU),
			Memory:    lipgloss.NewStyle().Foreground(styles.Zakura.Memory),
			Tokio:     lipgloss.NewStyle().Foreground(styles.Zakura.NodeActive),
			Component: lipgloss.NewStyle().Padding(0, 1),
			Separator: lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
		},
	}
	if source != nil {
		result.metrics = source.Snapshot().Metrics
	}
	return result
}

// ApplySnapshot records one dashboard sample.
func (m *Model) ApplySnapshot(metrics model.RuntimeMetrics, at time.Time) {
	current := sample{at: at, rss: metrics.Process.RSS, liveTasks: metrics.Rust.LiveTasks, hasRSS: metrics.Process.Has(model.ProcessMemory), hasTokio: metrics.Rust.TokioEnabled}
	if !m.sampledAt.IsZero() && at.After(m.sampledAt) && metrics.Process.Has(model.ProcessCPU) && m.metrics.Process.Has(model.ProcessCPU) {
		seconds := at.Sub(m.sampledAt).Seconds()
		currentCPU := metrics.Process.UserCPU + metrics.Process.SystemCPU
		previousCPU := m.metrics.Process.UserCPU + m.metrics.Process.SystemCPU
		if currentCPU >= previousCPU {
			current.cpu = (currentCPU - previousCPU) / seconds
			current.hasCPU = true
		}
	}
	m.metrics, m.sampledAt = metrics, at
	cutoff := at.Add(-historyWindow)
	first := 0
	for first < len(m.history) && m.history[first].at.Before(cutoff) {
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

// Init starts no work; the engine owns collection.
func (Model) Init() tea.Cmd { return nil }

// Update tracks terminal styling, dimensions, and compact-view scrolling.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
	case tea.BackgroundColorMsg:
		for _, style := range []*lipgloss.Style{&m.style.Root, &m.style.Section, &m.style.Label, &m.style.Value, &m.style.CPU, &m.style.Memory, &m.style.Tokio, &m.style.Component, &m.style.Separator} {
			*style = style.Background(msg.Color)
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			m.offset = max(0, m.offset-1)
		case "down", "j":
			m.offset++
		}
	}
	m.offset = min(m.offset, m.maxOffset())
	return m, nil
}

// Size returns the centered modal dimensions.
func Size(width, height int) (int, int) {
	return min(width, max(MinimumWidth, (width*ScreenPercent+50)/100)),
		min(height, max(MinimumHeight, (height*ScreenPercent+50)/100))
}

func (m Model) maxOffset() int {
	left := 3 + len(m.processRows()) + len(m.profileRows())
	right := 6 + len(m.tokioRows())
	return max(0, max(left, right)-max(0, m.height-2))
}
