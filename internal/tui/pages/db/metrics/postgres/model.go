// Package postgres renders PostgreSQL-specific database telemetry.
package postgres

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/styles"
)

const (
	minimumWidth  = 44
	minimumHeight = 10
	historyLimit  = 120
	historyWindow = 30 * time.Second
)

type sample struct {
	At              time.Time
	Transactions    float64
	Queries         float64
	Rows            float64
	BlockReads      float64
	BlockHits       float64
	TempBytes       float64
	Clients         uint64
	hasTransactions bool
	hasRows         bool
	hasBlocks       bool
	hasTemp         bool
	hasQueries      bool
}

// Model contains the latest PostgreSQL snapshot and bounded history.
type Model struct {
	metrics   model.PostgresMetrics
	history   []sample
	sampledAt time.Time
	style     Style
	width     int
	height    int
	offset    int
}

// Style contains dashboard presentation styles.
type Style struct {
	Root      lipgloss.Style
	Section   lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Activity  lipgloss.Style
	Storage   lipgloss.Style
	Column    lipgloss.Style
	Component lipgloss.Style
	Separator lipgloss.Style
}

// New creates a PostgreSQL dashboard.
func New(source *model.RuntimeGraph) Model {
	result := Model{
		history: make([]sample, 0, historyLimit),
		style: Style{
			Root:      styles.Panel(styles.Zakura).Padding(0),
			Section:   lipgloss.NewStyle().Foreground(styles.Zakura.Secondary).Bold(true),
			Label:     lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
			Value:     lipgloss.NewStyle().Foreground(styles.Zakura.Text),
			Activity:  lipgloss.NewStyle().Foreground(styles.Zakura.CPU),
			Storage:   lipgloss.NewStyle().Foreground(styles.Zakura.Memory),
			Column:    lipgloss.NewStyle(),
			Component: lipgloss.NewStyle().Padding(0, 1),
			Separator: lipgloss.NewStyle().Foreground(styles.Zakura.TextMuted),
		},
	}
	if source != nil {
		result.metrics = source.Snapshot().Metrics.Postgres
	}
	return result
}

// ApplySnapshot stores metrics and derives rates from cumulative counters.
func (m *Model) ApplySnapshot(metrics model.RuntimeMetrics, sampledAt time.Time) {
	currentMetrics := metrics.Postgres
	current := sample{At: sampledAt, Clients: currentMetrics.Backends}
	if !m.sampledAt.IsZero() && sampledAt.After(m.sampledAt) {
		seconds := sampledAt.Sub(m.sampledAt).Seconds()
		previous := m.metrics
		current.Transactions, current.hasTransactions = sumRate(
			[]uint64{currentMetrics.Commits, currentMetrics.Rollbacks},
			[]uint64{previous.Commits, previous.Rollbacks}, seconds,
		)
		current.Rows, current.hasRows = sumRate(
			[]uint64{currentMetrics.TuplesReturned, currentMetrics.TuplesFetched, currentMetrics.TuplesIn, currentMetrics.TuplesUpd, currentMetrics.TuplesDel},
			[]uint64{previous.TuplesReturned, previous.TuplesFetched, previous.TuplesIn, previous.TuplesUpd, previous.TuplesDel}, seconds,
		)
		current.BlockReads, current.hasBlocks = counterRate(currentMetrics.BlocksRead, previous.BlocksRead, seconds)
		var hitsOK bool
		current.BlockHits, hitsOK = counterRate(currentMetrics.BlocksHit, previous.BlocksHit, seconds)
		current.hasBlocks = current.hasBlocks && hitsOK
		current.TempBytes, current.hasTemp = counterRate(currentMetrics.TempBytes, previous.TempBytes, seconds)
		if currentMetrics.StatementsAvailable && previous.StatementsAvailable {
			current.Queries, current.hasQueries = counterRate(currentMetrics.StatementCalls, previous.StatementCalls, seconds)
		}
	}
	m.metrics = currentMetrics
	m.sampledAt = sampledAt
	m.appendHistory(current)
}

func counterRate(current, previous uint64, seconds float64) (float64, bool) {
	if seconds <= 0 || current < previous {
		return 0, false
	}
	return float64(current-previous) / seconds, true
}

func sumRate(current, previous []uint64, seconds float64) (float64, bool) {
	if len(current) != len(previous) || seconds <= 0 {
		return 0, false
	}
	var delta uint64
	for index := range current {
		if current[index] < previous[index] {
			return 0, false
		}
		delta += current[index] - previous[index]
	}
	return float64(delta) / seconds, true
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
			&m.style.Root, &m.style.Section, &m.style.Label, &m.style.Value,
			&m.style.Activity, &m.style.Storage, &m.style.Column,
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

func (m Model) maxOffset() int {
	left, right := m.compactColumns()
	return max(0, max(len(left), len(right))-max(0, m.height-2))
}

func (m Model) latestSample() sample {
	if len(m.history) == 0 {
		return sample{}
	}
	return m.history[len(m.history)-1]
}
