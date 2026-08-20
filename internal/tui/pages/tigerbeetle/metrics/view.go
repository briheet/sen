package metrics

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const metricLabelWidth = 15

// View renders the TigerBeetle dashboard.
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
	summaryHeight := max(10, (height-1)*45/100)
	historyHeight := max(0, height-1-summaryHeight)
	body := lipgloss.JoinVertical(lipgloss.Left,
		m.summary(width, summaryHeight),
		m.style.Separator.Render(strings.Repeat("─", width)),
		m.panel("HISTORY · 5M", m.historyRows(width-2, historyHeight-1), width, historyHeight),
	)
	return m.style.Root.Width(m.width).Height(m.height).Render(body)
}

func (m Model) summary(width, height int) string {
	columnWidth := max(0, (width-2)/3)
	lastWidth := max(0, width-2*columnWidth-2)
	divider := m.style.Separator.Render(strings.TrimSuffix(strings.Repeat("│\n", height), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel("CLUSTER", m.clusterRows(), columnWidth, height), divider,
		m.panel("ACTIVITY", m.activityRows(), columnWidth, height), divider,
		m.panel("STORAGE", m.storageRows(), lastWidth, height),
	)
}

func (m Model) compact() string {
	width := max(0, m.width-2)
	columnWidth := max(0, (width-4)/2)
	left, right := m.compactColumns()
	lines := make([]string, max(len(left), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, columnWidth) + strings.Repeat(" ", 4) +
			m.blockLine(right, index, max(0, width-columnWidth-4))
	}
	visible := max(0, m.height-2)
	start := min(m.offset, max(0, len(lines)-visible))
	end := min(len(lines), start+visible)
	return m.style.Root.Width(m.width).Height(m.height).Render(strings.Join(lines[start:end], "\n"))
}

func (m Model) compactColumns() ([]string, []string) {
	left := append([]string{m.style.Section.Render("CLUSTER")}, m.clusterRows()...)
	left = append(left, "", m.style.Section.Render("ACTIVITY"))
	left = append(left, m.activityRows()...)
	right := append([]string{m.style.Section.Render("STORAGE")}, m.storageRows()...)
	right = append(right, "", m.style.Section.Render("HISTORY · 5M"),
		m.historyRow("requests/s", requestsHistory, m.style.Activity),
		m.historyRow("latency", latencyHistory, m.style.Activity),
		m.historyRow("cache hit", cacheHistory, m.style.Storage),
	)
	return left, right
}

func (m Model) clusterRows() []string {
	healthy, stale, maxView, maxOperation, minCheckpoint := 0, 0, uint64(0), uint64(0), uint64(0)
	status := "waiting"
	for _, replica := range m.metrics.Replicas {
		if m.stale(replica) {
			stale++
			continue
		}
		if replica.Status == 0 {
			healthy++
		}
		maxView = max(maxView, replica.View)
		maxOperation = max(maxOperation, replica.Operation)
		if minCheckpoint == 0 || replica.Checkpoint < minCheckpoint {
			minCheckpoint = replica.Checkpoint
		}
	}
	if len(m.metrics.Replicas) > 0 {
		status = "healthy"
		if healthy != len(m.metrics.Replicas) || stale > 0 || len(m.metrics.Replicas) < m.expected {
			status = "degraded"
		}
	}
	statusStyle := m.style.Healthy
	if status != "healthy" {
		statusStyle = m.style.Warning
	}
	cluster := m.metrics.Cluster
	if cluster == "" {
		cluster = "waiting for StatsD"
	} else if len(cluster) > 16 {
		cluster = cluster[:16] + "…"
	}
	return []string{
		m.metric("cluster", cluster, m.style.Value),
		m.metric("release", availableCount(m.metrics.Release), m.style.Value),
		m.metric("status", status, statusStyle),
		m.metric("replicas", fmt.Sprintf("%d / %d", len(m.metrics.Replicas)-stale, m.expected), m.style.Value),
		m.metric("healthy", fmt.Sprintf("%d", healthy), m.style.Healthy),
		m.metric("view", fmt.Sprintf("%d", maxView), m.style.Value),
		m.metric("operation", fmt.Sprintf("%d", maxOperation), m.style.Value),
		m.metric("checkpoint", fmt.Sprintf("%d", minCheckpoint), m.style.Value),
	}
}

func (m Model) activityRows() []string {
	var requests uint64
	var latencySum time.Duration
	var latencyMax time.Duration
	var queue uint64
	busiest, busiestCount := "—", uint64(0)
	for name, operation := range m.metrics.Operations {
		requests += operation.Requests
		latencySum += operation.LatencySum
		latencyMax = max(latencyMax, operation.LatencyMax)
		if operation.Requests > busiestCount {
			busiest, busiestCount = name, operation.Requests
		}
	}
	for _, replica := range m.metrics.Replicas {
		queue = max(queue, replica.PipelineQueueLength)
	}
	average := time.Duration(0)
	if requests > 0 {
		average = latencySum / time.Duration(requests)
	}
	rate := float64(0)
	if seconds := m.metrics.Window.Seconds(); seconds > 0 {
		rate = float64(requests) / seconds
	}
	return []string{
		m.metric("requests/s", fmt.Sprintf("%.1f/s", rate), m.style.Activity),
		m.metric("requests window", fmt.Sprintf("%d", requests), m.style.Value),
		m.metric("latency avg", formatDuration(average), m.style.Activity),
		m.metric("latency max", formatDuration(latencyMax), m.style.Activity),
		m.metric("busiest", busiest, m.style.Value),
		m.metric("queue max", fmt.Sprintf("%d", queue), m.style.Value),
		m.metric("window", formatDuration(m.metrics.Window), m.style.Value),
	}
}

func (m Model) storageRows() []string {
	var hits, misses, missing, acquired, dirty, faulty, accounts, transfers uint64
	for _, replica := range m.metrics.Replicas {
		hits += replica.GridCacheHits
		misses += replica.GridCacheMisses
		missing += replica.GridBlocksMissing
		acquired += replica.GridBlocksAcquired
		dirty += replica.JournalDirty
		faulty += replica.JournalFaulty
		accounts = max(accounts, replica.Accounts)
		transfers = max(transfers, replica.Transfers)
	}
	cacheHit := "—"
	if total := hits + misses; total > 0 {
		cacheHit = fmt.Sprintf("%.1f%%", float64(hits)/float64(total)*100)
	}
	return []string{
		m.metric("cache hit", cacheHit, m.style.Storage),
		m.metric("hits / misses", fmt.Sprintf("%d / %d", hits, misses), m.style.Value),
		m.metric("blocks acquired", fmt.Sprintf("%d", acquired), m.style.Storage),
		m.metric("blocks missing", fmt.Sprintf("%d", missing), m.style.Warning),
		m.metric("journal dirty", fmt.Sprintf("%d", dirty), m.style.Warning),
		m.metric("journal faulty", fmt.Sprintf("%d", faulty), m.style.Warning),
		m.metric("accounts", fmt.Sprintf("%d", accounts), m.style.Value),
		m.metric("transfers", fmt.Sprintf("%d", transfers), m.style.Value),
	}
}

func (m Model) historyRows(width, height int) []string {
	columnWidth := max(1, (width-6)/3)
	lastWidth := max(1, width-2*columnWidth-6)
	blocks := [][]string{
		m.historyBlock("requests/s", requestsHistory, m.style.Activity, columnWidth, height),
		m.historyBlock("latency avg", latencyHistory, m.style.Activity, columnWidth, height),
		m.historyBlock("cache hit", cacheHistory, m.style.Storage, lastWidth, height),
	}
	lines := make([]string, max(len(blocks[0]), len(blocks[1]), len(blocks[2])))
	for index := range lines {
		lines[index] = m.blockLine(blocks[0], index, columnWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(blocks[1], index, columnWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(blocks[2], index, lastWidth)
	}
	return lines
}

func (m Model) historyBlock(name string, kind historyKind, style lipgloss.Style, width, height int) []string {
	value, ok := latestHistoryValue(m.history, kind)
	lines := []string{m.rowWidth(m.style.Label.Render(name), style.Render(formatHistory(value, kind, ok)), width)}
	return append(lines, historyChart(m.history, kind, width, max(1, height-1), style)...)
}

func (m Model) panel(title string, rows []string, width, height int) string {
	rows = rows[:min(len(rows), max(0, height-1))]
	return m.style.Component.Width(width).Height(height).Render(strings.Join(append([]string{m.style.Section.Render(title)}, rows...), "\n"))
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

func (m Model) rowWidth(left, right string, width int) string {
	return left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right))) + right
}

type historyKind uint8

const (
	requestsHistory historyKind = iota
	latencyHistory
	cacheHistory
)

func historyValue(sample sample, kind historyKind) float64 {
	switch kind {
	case requestsHistory:
		return sample.Requests
	case latencyHistory:
		return sample.Latency
	case cacheHistory:
		return sample.CacheHit
	default:
		return 0
	}
}

func (m Model) historyRow(label string, kind historyKind, style lipgloss.Style) string {
	columns := historyColumns(m.history, kind, 12)
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
	return m.metric(label, result.String(), style)
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
			current = historyColumn{value: historyValue(history[next], kind), available: true}
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

func latestHistoryValue(history []sample, kind historyKind) (float64, bool) {
	if len(history) == 0 {
		return 0, false
	}
	return historyValue(history[len(history)-1], kind), true
}

func formatHistory(value float64, kind historyKind, available bool) string {
	if !available {
		return "—"
	}
	switch kind {
	case requestsHistory:
		return fmt.Sprintf("%.1f/s", value)
	case latencyHistory:
		return formatDuration(time.Duration(value * float64(time.Microsecond)))
	case cacheHistory:
		return fmt.Sprintf("%.1f%%", value)
	default:
		return "—"
	}
}

func formatDuration(value time.Duration) string {
	if value == 0 {
		return "—"
	}
	return value.Round(time.Microsecond).String()
}

func availableCount(value uint64) string {
	if value == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", value)
}
