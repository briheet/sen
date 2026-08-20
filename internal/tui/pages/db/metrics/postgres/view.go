package postgres

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const metricLabelWidth = 16

// View renders a full dashboard when space permits and a scrollable compact
// view in smaller terminals.
func (m Model) View() string {
	if m.width < minimumWidth || m.height < minimumHeight {
		return ""
	}
	if m.width >= 96 && m.height >= 28 {
		return m.dashboard()
	}
	return m.compact()
}

func (m Model) dashboard() string {
	width, height := max(0, m.width-2), max(0, m.height-2)
	summaryHeight, historyHeight := dashboardHeights(height)
	body := lipgloss.JoinVertical(lipgloss.Left,
		m.summary(width, summaryHeight),
		m.style.Separator.Render(strings.Repeat("─", width)),
		m.panel("HISTORY · 30S", m.historyRows(width-2, historyHeight-1), width, historyHeight),
	)
	return m.style.Root.Width(m.width).Height(m.height).Render(body)
}

func dashboardHeights(height int) (int, int) {
	content := max(0, height-1)
	summary := max(10, content*45/100)
	return summary, max(0, content-summary)
}

func (m Model) summary(width, height int) string {
	columnWidth := max(0, (width-2)/3)
	lastWidth := max(0, width-2*columnWidth-2)
	divider := m.style.Separator.Render(strings.TrimSuffix(strings.Repeat("│\n", height), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel("DATABASE", m.databaseRows(), columnWidth, height), divider,
		m.panel("ACTIVITY", m.activityRows(), columnWidth, height), divider,
		m.panel("STORAGE", m.storageRows(), lastWidth, height),
	)
}

func (m Model) compact() string {
	width := max(0, m.width-2)
	columnWidth := max(0, (width-4)/2)
	left, right := m.compactColumns()
	body := m.columns(left, right, width, columnWidth)
	visible := max(0, m.height-2)
	start := min(m.offset, max(0, len(body)-visible))
	end := min(len(body), start+visible)
	return m.style.Root.Width(m.width).Height(m.height).Render(strings.Join(body[start:end], "\n"))
}

func (m Model) compactColumns() ([]string, []string) {
	left := append([]string{m.style.Section.Render("DATABASE")}, m.databaseRows()...)
	left = append(left, "", m.style.Section.Render("ACTIVITY"))
	left = append(left, m.activityRows()...)
	right := append([]string{m.style.Section.Render("STORAGE")}, m.storageRows()...)
	right = append(right, "", m.style.Section.Render("HISTORY · 30S"),
		m.historyRow("transactions", transactionsHistory, m.style.Activity),
		m.historyRow("queries", queriesHistory, m.style.Activity),
		m.historyRow("clients", clientsHistory, m.style.Value),
	)
	return left, right
}

func (m Model) databaseRows() []string {
	metrics := m.metrics
	return []string{
		m.metricAvailable("version", metrics.Version, metrics.Version != "", m.style.Value),
		m.metricAvailable("database", metrics.Database, metrics.Database != "", m.style.Value),
		m.metric("uptime", formatDuration(metrics.Uptime), m.style.Value),
		m.metric("size", formatBytes(metrics.DatabaseSize), m.style.Storage),
		m.metric("connections", fmt.Sprintf("%d / %s", metrics.Backends, formatLimit(metrics.MaxConnections)), m.style.Value),
		m.metric("active", formatCount(metrics.Active), m.style.Activity),
		m.metric("idle", formatCount(metrics.Idle), m.style.Value),
		m.metric("waiting", formatCount(metrics.Waiting), m.style.Activity),
		m.metric("locks", formatCount(metrics.Locks), m.style.Value),
	}
}

func (m Model) activityRows() []string {
	metrics, latest := m.metrics, m.latestSample()
	return []string{
		m.metricAvailable("transactions/s", formatRate(latest.Transactions), latest.hasTransactions, m.style.Activity),
		m.metricAvailable("queries/s", formatRate(latest.Queries), latest.hasQueries, m.style.Activity),
		m.metric("commits total", formatCount(metrics.Commits), m.style.Value),
		m.metric("rollbacks total", formatCount(metrics.Rollbacks), m.style.Value),
		m.metric("statement calls", formatAvailableCount(metrics.StatementCalls, metrics.StatementsAvailable), m.style.Value),
		m.metricAvailable("rows/s", formatRate(latest.Rows), latest.hasRows, m.style.Value),
		m.metric("deadlocks", formatCount(metrics.Deadlocks), m.style.Activity),
	}
}

func (m Model) storageRows() []string {
	metrics, latest := m.metrics, m.latestSample()
	return []string{
		m.metric("cache hit", cacheHitRatio(metrics.BlocksHit, metrics.BlocksRead), m.style.Storage),
		m.metricAvailable("blocks read/s", formatRate(latest.BlockReads), latest.hasBlocks, m.style.Value),
		m.metricAvailable("blocks hit/s", formatRate(latest.BlockHits), latest.hasBlocks, m.style.Value),
		m.metric("blocks read", formatCount(metrics.BlocksRead), m.style.Value),
		m.metric("blocks hit", formatCount(metrics.BlocksHit), m.style.Value),
		m.metricAvailable("temp write/s", formatByteRate(latest.TempBytes), latest.hasTemp, m.style.Storage),
		m.metric("temp bytes", formatBytes(metrics.TempBytes), m.style.Storage),
		m.metric("temp files", formatCount(metrics.TempFiles), m.style.Value),
	}
}

func (m Model) historyRows(width, height int) []string {
	columnWidth := max(1, (width-6)/3)
	lastWidth := max(1, width-2*columnWidth-6)
	left := m.historyBlock("transactions/s", transactionsHistory, m.style.Activity, columnWidth, height)
	middle := m.historyBlock("queries/s", queriesHistory, m.style.Activity, columnWidth, height)
	right := m.historyBlock("clients", clientsHistory, m.style.Value, lastWidth, height)
	lines := make([]string, max(len(left), len(middle), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, columnWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(middle, index, columnWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(right, index, lastWidth)
	}
	return lines
}

func (m Model) historyBlock(name string, kind historyKind, style lipgloss.Style, width, height int) []string {
	value, ok := latestHistoryValue(m.history, kind)
	lines := []string{m.rowWidth(m.style.Label.Render(name), style.Render(formatHistoryValue(value, kind, ok)), width)}
	return append(lines, historyChart(m.history, kind, width, max(1, height-1), style)...)
}

func (m Model) panel(title string, rows []string, width, height int) string {
	rows = rows[:min(len(rows), max(0, height-1))]
	content := append([]string{m.style.Section.Render(title)}, rows...)
	return m.style.Component.Width(width).Height(height).Render(strings.Join(content, "\n"))
}

func (m Model) columns(left, right []string, width, columnWidth int) []string {
	lines := make([]string, max(len(left), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, columnWidth) + strings.Repeat(" ", 4) +
			m.blockLine(right, index, max(0, width-columnWidth-4))
	}
	return lines
}

func (m Model) blockLine(lines []string, index, width int) string {
	if index >= len(lines) {
		return strings.Repeat(" ", max(0, width))
	}
	return m.style.Column.Width(width).Render(ansi.Truncate(lines[index], width, ""))
}

func (m Model) metric(label, value string, style lipgloss.Style) string {
	label += strings.Repeat(" ", max(0, metricLabelWidth-lipgloss.Width(label))) + ":"
	return m.style.Label.Render(label) + " " + style.Render(value)
}

func (m Model) metricAvailable(label, value string, available bool, style lipgloss.Style) string {
	if !available {
		value = "—"
	}
	return m.metric(label, value, style)
}

func (m Model) rowWidth(left, right string, width int) string {
	return left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right))) + right
}

type historyKind uint8

const (
	transactionsHistory historyKind = iota
	queriesHistory
	clientsHistory
)

func (m Model) historyRow(label string, kind historyKind, style lipgloss.Style) string {
	return m.metric(label, historySparkline(m.history, kind, 12), style)
}

func historyValue(sample sample, kind historyKind) (float64, bool) {
	switch kind {
	case transactionsHistory:
		return sample.Transactions, sample.hasTransactions
	case queriesHistory:
		return sample.Queries, sample.hasQueries
	case clientsHistory:
		return float64(sample.Clients), true
	default:
		return 0, false
	}
}

func latestHistoryValue(history []sample, kind historyKind) (float64, bool) {
	if len(history) == 0 {
		return 0, false
	}
	return historyValue(history[len(history)-1], kind)
}

func historySparkline(history []sample, kind historyKind, width int) string {
	columns := historyColumns(history, kind, width)
	peak := historyPeak(columns)
	levels := []rune("▁▂▃▄▅▆▇█")
	var result strings.Builder
	for _, column := range columns {
		if !column.available {
			result.WriteRune('·')
			continue
		}
		level := 0
		if peak > 0 {
			level = int(column.value * float64(len(levels)-1) / peak)
		}
		result.WriteRune(levels[level])
	}
	return result.String()
}

func historyChart(history []sample, kind historyKind, width, height int, style lipgloss.Style) []string {
	columns := historyColumns(history, kind, width)
	peak := historyPeak(columns)
	units := make([]int, len(columns))
	for index, column := range columns {
		if !column.available {
			continue
		}
		units[index] = 1
		if peak > 0 {
			units[index] = max(1, int(math.Ceil(column.value*float64(height*8)/peak)))
		}
	}
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	lines := make([]string, max(0, height))
	for row := range lines {
		var line strings.Builder
		for _, value := range units {
			line.WriteRune(blocks[min(8, max(0, value-(height-row-1)*8))])
		}
		lines[row] = style.Render(line.String())
	}
	return lines
}

type historyColumn struct {
	value     float64
	available bool
}

func historyColumns(history []sample, kind historyKind, width int) []historyColumn {
	columns := make([]historyColumn, max(0, width))
	if len(history) == 0 || width <= 0 {
		return columns
	}
	windowEnd := history[len(history)-1].At
	windowStart := windowEnd.Add(-historyWindow)
	next := 0
	current := historyColumn{}
	for index := range columns {
		at := windowStart.Add(time.Duration(index+1) * historyWindow / time.Duration(width))
		for next < len(history) && !history[next].At.After(at) {
			current.value, current.available = historyValue(history[next], kind)
			next++
		}
		columns[index] = current
	}
	return columns
}

func historyPeak(columns []historyColumn) float64 {
	peak := 0.0
	for _, column := range columns {
		if column.available {
			peak = max(peak, column.value)
		}
	}
	return peak
}

func formatHistoryValue(value float64, kind historyKind, ok bool) string {
	if !ok {
		return "—"
	}
	if kind == clientsHistory {
		return formatCount(uint64(value))
	}
	return formatRate(value)
}

func cacheHitRatio(hits, reads uint64) string {
	if total := hits + reads; total > 0 {
		return fmt.Sprintf("%.1f%%", float64(hits)/float64(total)*100)
	}
	return "—"
}

func formatDuration(value time.Duration) string { return value.Round(time.Second).String() }

func formatBytes(value uint64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}
	divisor, exponent := uint64(unit), 0
	for next := value / unit; next >= unit && exponent < 5; next /= unit {
		divisor *= unit
		exponent++
	}
	return fmt.Sprintf("%.1f %ciB", float64(value)/float64(divisor), "KMGTPE"[exponent])
}

func formatCount(value uint64) string     { return fmt.Sprintf("%d", value) }
func formatRate(value float64) string     { return fmt.Sprintf("%.1f/s", value) }
func formatByteRate(value float64) string { return formatBytes(uint64(max(0, value))) + "/s" }
func formatLimit(value uint64) string {
	if value == 0 {
		return "—"
	}
	return formatCount(value)
}
func formatAvailableCount(value uint64, available bool) string {
	if !available {
		return "—"
	}
	return formatCount(value)
}
