// Package metrics renders process telemetry for a server page.
package metrics

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
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

// Model contains the runtime snapshot rendered by the metrics overlay.
type Model struct {
	metrics   model.RuntimeMetrics
	history   []sample
	sampledAt time.Time
	scheduler histogramWindow
	gcPauses  histogramWindow
	lang      config.ServiceLang
	style     Style
	width     int
	height    int
	offset    int
}

type sample struct {
	At             time.Time
	CPUCores       float64
	AllocationRate float64
	AllocationOps  float64
	GCCycleRate    float64
	GCAssist       float64
	ReadRate       float64
	WriteRate      float64
	ReadOps        float64
	WriteOps       float64
	ContextRate    float64
	LiveMemory     uint64
	RSS            uint64
	EventLoopP99   time.Duration
	available      sampleMetric
}

type sampleMetric uint16

const (
	sampleCPU sampleMetric = 1 << iota
	sampleAllocation
	sampleAllocationOps
	sampleGC
	sampleGCAssist
	sampleIO
	sampleIOOps
	sampleContext
	sampleRSS
)

// Style contains metrics panel presentation styles.
type Style struct {
	Root      lipgloss.Style
	Section   lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	CPU       lipgloss.Style
	Memory    lipgloss.Style
	Column    lipgloss.Style
	Component lipgloss.Style
	Separator lipgloss.Style
}

// New creates a metrics panel backed by the engine's runtime graph.
func New(source *model.RuntimeGraph, lang config.ServiceLang) Model {
	result := Model{
		lang:    lang,
		history: make([]sample, 0, historyLimit),
		style: Style{
			Root:      styles.Panel(styles.Zakura).Padding(0),
			Section:   lipgloss.NewStyle().Foreground(styles.Zakura.Secondary).Bold(true),
			Label:     lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
			Value:     lipgloss.NewStyle().Foreground(styles.Zakura.Text),
			CPU:       lipgloss.NewStyle().Foreground(styles.Zakura.CPU),
			Memory:    lipgloss.NewStyle().Foreground(styles.Zakura.Memory),
			Column:    lipgloss.NewStyle(),
			Component: lipgloss.NewStyle().Padding(0, 1),
			Separator: lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
		},
	}
	if source != nil {
		result.metrics = source.Snapshot().Metrics
	}
	return result
}

// ApplySnapshot stores the latest metrics and appends one bounded history point.
func (m *Model) ApplySnapshot(metrics model.RuntimeMetrics, sampledAt time.Time) {
	current := sample{At: sampledAt, RSS: metrics.Process.RSS, EventLoopP99: metrics.Node.EventLoopDelayP99}
	if metrics.Process.Has(model.ProcessMemory) {
		current.available |= sampleRSS
	}
	if m.lang == config.ServiceLangNode {
		current.LiveMemory = metrics.Node.HeapUsed
	} else {
		current.LiveMemory = metrics.Go.LiveHeap
	}
	if !m.sampledAt.IsZero() && sampledAt.After(m.sampledAt) {
		seconds := sampledAt.Sub(m.sampledAt).Seconds()
		if metrics.Process.Has(model.ProcessCPU) && m.metrics.Process.Has(model.ProcessCPU) {
			currentCPU := metrics.Process.UserCPU + metrics.Process.SystemCPU
			previousCPU := m.metrics.Process.UserCPU + m.metrics.Process.SystemCPU
			if currentCPU >= previousCPU {
				current.CPUCores = (currentCPU - previousCPU) / seconds
				current.available |= sampleCPU
			}
		}
		if metrics.Process.Has(model.ProcessIO) && m.metrics.Process.Has(model.ProcessIO) &&
			metrics.Process.ReadBytes >= m.metrics.Process.ReadBytes && metrics.Process.WriteBytes >= m.metrics.Process.WriteBytes {
			current.ReadRate = float64(metrics.Process.ReadBytes-m.metrics.Process.ReadBytes) / seconds
			current.WriteRate = float64(metrics.Process.WriteBytes-m.metrics.Process.WriteBytes) / seconds
			current.available |= sampleIO
		}
		if metrics.Process.Has(model.ProcessIOOperations) && m.metrics.Process.Has(model.ProcessIOOperations) &&
			metrics.Process.ReadOps >= m.metrics.Process.ReadOps && metrics.Process.WriteOps >= m.metrics.Process.WriteOps {
			current.ReadOps = float64(metrics.Process.ReadOps-m.metrics.Process.ReadOps) / seconds
			current.WriteOps = float64(metrics.Process.WriteOps-m.metrics.Process.WriteOps) / seconds
			current.available |= sampleIOOps
		}
		if metrics.Process.Has(model.ProcessContextSwitches) && m.metrics.Process.Has(model.ProcessContextSwitches) {
			currentTotal := metrics.Process.VoluntaryCS + metrics.Process.InvoluntaryCS
			previousTotal := m.metrics.Process.VoluntaryCS + m.metrics.Process.InvoluntaryCS
			if currentTotal >= previousTotal {
				current.ContextRate = float64(currentTotal-previousTotal) / seconds
				current.available |= sampleContext
			}
		}
		if m.lang != config.ServiceLangNode && metrics.Go.AllocatedBytes >= m.metrics.Go.AllocatedBytes {
			current.AllocationRate = float64(metrics.Go.AllocatedBytes-m.metrics.Go.AllocatedBytes) / seconds
			current.available |= sampleAllocation
		}
		if m.lang != config.ServiceLangNode && metrics.Go.Allocations >= m.metrics.Go.Allocations {
			current.AllocationOps = float64(metrics.Go.Allocations-m.metrics.Go.Allocations) / seconds
			current.available |= sampleAllocationOps
		}
		if m.lang != config.ServiceLangNode && metrics.Go.GCCycles >= m.metrics.Go.GCCycles {
			current.GCCycleRate = float64(metrics.Go.GCCycles-m.metrics.Go.GCCycles) / seconds
			current.available |= sampleGC
		}
		if m.lang != config.ServiceLangNode && metrics.Go.GCAssist >= m.metrics.Go.GCAssist {
			current.GCAssist = (metrics.Go.GCAssist - m.metrics.Go.GCAssist) / seconds * 100
			current.available |= sampleGCAssist
		}
	}
	m.scheduler.Add(sampledAt, metrics.Go.SchedulerLatency)
	m.gcPauses.Add(sampledAt, metrics.Go.GCPauses)
	m.metrics = metrics
	m.sampledAt = sampledAt
	cutoff := sampledAt.Add(-historyWindow)
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

// Init starts no background work; the engine owns metric collection.
func (Model) Init() tea.Cmd { return nil }

// Update tracks the overlay dimensions through Bubble Tea messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(0, msg.Width)
		m.height = max(0, msg.Height)
	case tea.BackgroundColorMsg:
		// Match the terminal while making overlay whitespace opaque.
		for _, style := range []*lipgloss.Style{
			&m.style.Root, &m.style.Section, &m.style.Label,
			&m.style.Value, &m.style.CPU, &m.style.Memory,
			&m.style.Column, &m.style.Component, &m.style.Separator,
		} {
			*style = style.Background(msg.Color).ColorWhitespace(true)
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
	return min(width, max(MinimumWidth, (width*ScreenPercent+50)/100)),
		min(height, max(MinimumHeight, (height*ScreenPercent+50)/100))
}

func (m Model) maxOffset() int {
	contentHeight := 23
	if m.lang == config.ServiceLangNode {
		contentHeight = 19
	}
	return max(0, contentHeight-max(0, m.height-2))
}
