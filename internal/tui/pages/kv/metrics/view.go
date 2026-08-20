package metrics

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const metricLabelWidth = 16

// View renders the Redis dashboard at the current modal size.
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
	width := max(0, m.width-2)
	height := max(0, m.height-2)
	summaryHeight, historyHeight := dashboardHeights(height)
	summary := m.summary(width, summaryHeight)
	history := m.panel("HISTORY · 30S", m.historyRows(width-2, historyHeight-1), width, historyHeight)
	body := lipgloss.JoinVertical(lipgloss.Left,
		summary,
		m.style.Separator.Render(strings.Repeat("─", width)),
		history,
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
		m.panel("SERVER", m.serverRows(), columnWidth, height), divider,
		m.panel("MEMORY", m.memoryRows(), columnWidth, height), divider,
		m.panel("ACTIVITY", m.activityRows(), lastWidth, height),
	)
}

func (m Model) compact() string {
	width := max(0, m.width-2)
	columnWidth := max(0, (width-4)/2)
	left := append([]string{m.style.Section.Render("SERVER")}, m.serverRows()...)
	left = append(left, "", m.style.Section.Render("ACTIVITY"))
	left = append(left, m.activityRows()...)
	right := append([]string{m.style.Section.Render("MEMORY")}, m.memoryRows()...)
	right = append(right, "", m.style.Section.Render("HISTORY · 30S"),
		m.historyRow("ops/s", operationsHistory, m.style.Activity),
		m.historyRow("used", memoryHistory, m.style.Memory),
		m.historyRow("clients", clientsHistory, m.style.Value),
	)
	body := m.columns(left, right, width, columnWidth)
	visible := max(0, m.height-2)
	start := min(m.offset, max(0, len(body)-visible))
	end := min(len(body), start+visible)
	return m.style.Root.Width(m.width).Height(m.height).Render(strings.Join(body[start:end], "\n"))
}

func (m Model) serverRows() []string {
	metrics := m.metrics
	return []string{
		m.metricAvailable("version", metrics.Version, metrics.Version != "", m.style.Value),
		m.metricAvailable("mode", metrics.Mode, metrics.Mode != "", m.style.Value),
		m.metricAvailable("role", metrics.Role, metrics.Role != "", m.style.Value),
		m.metric("uptime", formatDuration(metrics.Uptime), m.style.Value),
		m.metric("keys", formatCount(metrics.Keys), m.style.Value),
		m.metric("clients", formatCount(metrics.ConnectedClients), m.style.Value),
		m.metric("blocked", formatCount(metrics.BlockedClients), m.style.Value),
		m.metric("cpu user / sys", formatCPU(metrics.UserCPU)+" / "+formatCPU(metrics.SystemCPU), m.style.Value),
	}
}

func (m Model) memoryRows() []string {
	metrics := m.metrics
	maxMemory := "unlimited"
	if metrics.MaxMemory > 0 {
		maxMemory = formatBytes(metrics.MaxMemory)
	}
	return []string{
		m.metric("used", formatBytes(metrics.UsedMemory), m.style.Memory),
		m.metric("dataset", formatBytes(metrics.UsedMemoryDataset), m.style.Memory),
		m.metric("rss", formatBytes(metrics.RSS), m.style.Memory),
		m.metric("peak", formatBytes(metrics.PeakMemory), m.style.Memory),
		m.metric("maxmemory", maxMemory, m.style.Memory),
		m.metric("fragmentation", fmt.Sprintf("%.2fx", metrics.MemoryFragmentationRatio), m.style.Memory),
		m.metricAvailable("dataset usage", formatRatio(metrics.UsedMemoryDataset, metrics.UsedMemory), metrics.UsedMemory > 0, m.style.Memory),
	}
}

func (m Model) activityRows() []string {
	metrics := m.metrics
	latest := m.latestSample()
	cacheRatio := cumulativeCacheRatio(metrics.KeyspaceHits, metrics.KeyspaceMisses)
	if latest.hasCacheRatio {
		cacheRatio = fmt.Sprintf("%.1f%% window", latest.CacheHitRatio)
	}
	return []string{
		m.metric("commands/s", formatRate(float64(metrics.InstantaneousOps)), m.style.Activity),
		m.metric("commands total", formatCount(metrics.TotalCommandsProcessed), m.style.Activity),
		m.metricAvailable("connections/s", formatRate(latest.Connections), latest.hasConnectionRate, m.style.Value),
		m.metricAvailable("network in", formatByteRate(latest.NetworkInput), latest.hasInputRate, m.style.Value),
		m.metricAvailable("network out", formatByteRate(latest.NetworkOutput), latest.hasOutputRate, m.style.Value),
		m.metric("cache hit", cacheRatio, m.style.Activity),
		m.metric("hits / misses", formatCount(metrics.KeyspaceHits)+" / "+formatCount(metrics.KeyspaceMisses), m.style.Value),
		m.metric("expired", formatCount(metrics.ExpiredKeys), m.style.Value),
		m.metric("evicted", formatCount(metrics.EvictedKeys), m.style.Value),
	}
}

func cumulativeCacheRatio(hits, misses uint64) string {
	if total := hits + misses; total > 0 {
		return fmt.Sprintf("%.1f%%", float64(hits)/float64(total)*100)
	}
	return "—"
}

func (m Model) historyRows(width, height int) []string {
	columnWidth := max(1, (width-6)/3)
	lastWidth := max(1, width-2*columnWidth-6)
	left := m.historyBlock("commands/s", operationsHistory, m.style.Activity, columnWidth, height)
	middle := m.historyBlock("used memory", memoryHistory, m.style.Memory, columnWidth, height)
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
	visible := max(0, height-1)
	rows = rows[:min(len(rows), visible)]
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
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

type historyKind uint8

const (
	operationsHistory historyKind = iota
	memoryHistory
	clientsHistory
)

func (m Model) historyRow(label string, kind historyKind, style lipgloss.Style) string {
	return m.metric(label, historySparkline(m.history, kind, 12), style)
}

func historyValue(sample sample, kind historyKind) float64 {
	switch kind {
	case operationsHistory:
		return sample.Operations
	case memoryHistory:
		return float64(sample.UsedMemory)
	case clientsHistory:
		return float64(sample.Clients)
	default:
		return 0
	}
}

func latestHistoryValue(history []sample, kind historyKind) (float64, bool) {
	if len(history) == 0 {
		return 0, false
	}
	return historyValue(history[len(history)-1], kind), true
}

func historySparkline(history []sample, kind historyKind, width int) string {
	columns := historyColumns(history, kind, width)
	peak := 0.0
	for _, column := range columns {
		if column.available {
			peak = max(peak, column.value)
		}
	}
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
	peak := 0.0
	for _, column := range columns {
		if column.available {
			peak = max(peak, column.value)
		}
	}
	units := make([]int, len(columns))
	for index, column := range columns {
		if !column.available {
			continue
		}
		units[index] = 1 // Keep zero-valued samples visible on the time axis.
		if peak > 0 {
			units[index] = max(1, int(math.Ceil(column.value*float64(height*8)/peak)))
		}
	}
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	lines := make([]string, max(0, height))
	for row := range lines {
		var line strings.Builder
		for _, value := range units {
			level := min(8, max(0, value-(height-row-1)*8))
			line.WriteRune(blocks[level])
		}
		lines[row] = style.Render(line.String())
	}
	return lines
}

type historyColumn struct {
	value     float64
	available bool
}

// historyColumns projects the fixed time window onto every cell in the plot.
// A sample remains active until the next sample, matching a metrics step chart.
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

func formatHistoryValue(value float64, kind historyKind, ok bool) string {
	if !ok {
		return "—"
	}
	switch kind {
	case operationsHistory:
		return formatRate(value)
	case memoryHistory:
		return formatBytes(uint64(value))
	case clientsHistory:
		return formatCount(uint64(value))
	default:
		return "—"
	}
}

func formatDuration(duration time.Duration) string {
	if duration > 0 && duration < time.Millisecond {
		return duration.Round(time.Microsecond).String()
	}
	return duration.Round(time.Millisecond).String()
}

func formatCPU(seconds float64) string {
	return formatDuration(time.Duration(seconds * float64(time.Second)))
}

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

func formatCount(value uint64) string { return fmt.Sprintf("%d", value) }

func formatRate(value float64) string { return fmt.Sprintf("%.1f/s", value) }

func formatByteRate(value float64) string { return formatBytes(uint64(max(0, value))) + "/s" }

func formatRatio(value, total uint64) string {
	if total == 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", float64(value)/float64(total)*100)
}
