package metrics

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/x/ansi"
)

const metricLabelWidth = 16

// View renders every metric available for the selected runtime.
func (m Model) View() string {
	if m.width < MinimumWidth || m.height < MinimumHeight {
		return ""
	}
	if m.width >= 96 && m.height >= 28 {
		return m.dashboard()
	}
	return m.compact()
}

// dashboard keeps summaries above the denser rolling telemetry.
func (m Model) dashboard() string {
	width := max(0, m.width-2)
	height := max(0, m.height-2)
	body := m.goDashboard(width, height)
	if m.lang == config.ServiceLangNode {
		body = m.nodeDashboard(width, height)
	}
	return m.style.Root.Width(m.width).Height(m.height).Render(body)
}

func (m Model) goDashboard(width, height int) string {
	summaryHeight, detailHeight := dashboardHeights(height)
	summary := m.summary(
		[]string{"PROCESS", "GO RUNTIME", "MEMORY"},
		[][]string{m.processRows(), m.goRuntimeRows(), m.goMemoryRows()},
		width, summaryHeight,
	)
	details := m.panel("HISTOGRAMS · 30S", m.histogramRows(width-2, detailHeight-1), width, detailHeight)
	return lipgloss.JoinVertical(lipgloss.Left,
		summary,
		m.style.Separator.Render(strings.Repeat("─", width)),
		details,
	)
}

func (m Model) nodeDashboard(width, height int) string {
	summaryHeight, detailHeight := dashboardHeights(height)
	summary := m.summary(
		[]string{"PROCESS", "EVENT LOOP", "MEMORY"},
		[][]string{m.processRows(), m.nodeEventRows(), m.nodeMemoryRows()},
		width, summaryHeight,
	)
	details := m.panel("HISTORY · 30S", m.historyRows(width-2, detailHeight-1), width, detailHeight)
	return lipgloss.JoinVertical(lipgloss.Left,
		summary,
		m.style.Separator.Render(strings.Repeat("─", width)),
		details,
	)
}

func dashboardHeights(height int) (int, int) {
	content := max(0, height-1) // One row separates the two sections.
	summary := max(10, content*45/100)
	return summary, max(0, content-summary)
}

func (m Model) summary(titles []string, rows [][]string, width, height int) string {
	columnWidth := max(0, (width-2)/3)
	lastWidth := max(0, width-2*columnWidth-2)
	divider := m.style.Separator.Render(strings.TrimSuffix(strings.Repeat("│\n", height), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel(titles[0], rows[0], columnWidth, height), divider,
		m.panel(titles[1], rows[1], columnWidth, height), divider,
		m.panel(titles[2], rows[2], lastWidth, height),
	)
}

func (m Model) compact() string {
	width := max(0, m.width-2)
	columnWidth := max(0, (width-4)/2)
	left := append([]string{m.style.Section.Render("PROCESS")}, m.processRows()...)
	var middle, right []string
	if m.lang == config.ServiceLangNode {
		middle = append([]string{m.style.Section.Render("EVENT LOOP")}, m.nodeEventRows()...)
		right = append([]string{m.style.Section.Render("MEMORY")}, m.nodeMemoryRows()...)
	} else {
		middle = append([]string{m.style.Section.Render("GO RUNTIME")}, m.goRuntimeRows()...)
		right = append([]string{m.style.Section.Render("MEMORY")}, m.goMemoryRows()...)
	}
	left = append(left, "")
	left = append(left, middle...)
	if m.lang == config.ServiceLangNode {
		right = append(right, "", m.style.Section.Render("HISTORY · 30S"),
			m.historyRow("cpu", cpuHistory, m.style.CPU),
			m.historyRow("rss", rssHistory, m.style.Memory),
			m.historyRow("event p99", eventLoopHistory, m.style.CPU),
		)
	} else {
		right = append(right, "", m.style.Section.Render("HISTOGRAMS · 30S"),
			m.metric("scheduler p95", formatQuantile(m.scheduler.Histogram(), 0.95), m.style.CPU),
			m.metric("scheduler p99", formatQuantile(m.scheduler.Histogram(), 0.99), m.style.CPU),
			m.metric("gc p95", formatQuantile(m.gcPauses.Histogram(), 0.95), m.style.Memory),
			m.metric("gc p99", formatQuantile(m.gcPauses.Histogram(), 0.99), m.style.Memory),
		)
	}
	body := m.columns(left, right, width, columnWidth)
	visible := max(0, m.height-2)
	start := min(m.offset, max(0, len(body)-visible))
	end := min(len(body), start+visible)
	return m.style.Root.Width(m.width).Height(m.height).Render(strings.Join(body[start:end], "\n"))
}

func (m Model) panel(title string, rows []string, width, height int) string {
	visible := max(0, height-1)
	rows = rows[:min(len(rows), visible)]
	content := append([]string{m.style.Section.Render(title)}, rows...)
	return m.style.Component.Width(width).Height(height).Render(strings.Join(content, "\n"))
}

func (m Model) processRows() []string {
	metrics := m.metrics.Process
	latest := m.latestSample()
	return []string{
		m.metricAvailable("uptime", formatDuration(metrics.Uptime), metrics.Has(model.ProcessUptime), m.style.Value),
		m.metricAvailable("cpu", fmt.Sprintf("%.2f cores", latest.CPUCores), latest.available&sampleCPU != 0, m.style.CPU),
		m.metricAvailable("cpu time u/s", formatSeconds(metrics.UserCPU)+" / "+formatSeconds(metrics.SystemCPU), metrics.Has(model.ProcessCPU), m.style.CPU),
		m.metricAvailable("threads", formatCount(metrics.Threads), metrics.Has(model.ProcessThreads), m.style.Value),
		m.metricAvailable("context switches", formatRate(latest.ContextRate), latest.available&sampleContext != 0, m.style.Value),
		m.metricAvailable("rss", formatBytes(metrics.RSS), metrics.Has(model.ProcessMemory), m.style.Memory),
		m.metricAvailable("peak rss", formatBytes(metrics.PeakRSS), metrics.Has(model.ProcessMemory), m.style.Memory),
		m.metricAvailable("open files", formatCount(metrics.OpenFiles), metrics.Has(model.ProcessOpenFiles), m.style.Value),
		m.metricAvailable("io bytes r/w", formatByteRate(latest.ReadRate)+" / "+formatByteRate(latest.WriteRate), latest.available&sampleIO != 0, m.style.Value),
		m.metricAvailable("io ops r/w", formatRate(latest.ReadOps)+" / "+formatRate(latest.WriteOps), latest.available&sampleIOOps != 0, m.style.Value),
	}
}

func (m Model) goRuntimeRows() []string {
	metrics := m.metrics.Go
	latest := m.latestSample()
	return []string{
		m.metric("goroutines", formatCount(metrics.Goroutines), m.style.Value),
		m.metric("GOMAXPROCS", formatCount(metrics.GOMAXPROCS), m.style.Value),
		m.metric("runtime cpu", formatSeconds(metrics.UserCPU), m.style.CPU),
		m.metricAvailable("alloc bytes", formatByteRate(latest.AllocationRate), latest.available&sampleAllocation != 0, m.style.Value),
		m.metricAvailable("alloc objects", formatRate(latest.AllocationOps), latest.available&sampleAllocationOps != 0, m.style.Value),
		m.metricAvailable("gc rate", formatRate(latest.GCCycleRate), latest.available&sampleGC != 0, m.style.CPU),
		m.metricAvailable("gc assist", formatPercent(latest.GCAssist), latest.available&sampleGCAssist != 0, m.style.CPU),
		m.metric("gc cpu", formatSeconds(metrics.GCCPU), m.style.CPU),
		m.metric("mutex wait", formatSeconds(metrics.MutexWait), m.style.Value),
		m.metric("GOGC", formatGOGC(metrics.GOGC), m.style.Value),
	}
}

func (m Model) goMemoryRows() []string {
	metrics := m.metrics.Go
	process := m.metrics.Process
	managedResident := metrics.RuntimeMemory - min(metrics.RuntimeMemory, metrics.HeapReleased)
	rssGap := process.RSS - min(process.RSS, managedResident)
	return []string{
		m.metric("live heap", formatBytes(metrics.LiveHeap), m.style.Memory),
		m.metric("heap goal", formatBytes(metrics.HeapGoal), m.style.Memory),
		m.metricAvailable("heap usage", formatRatio(metrics.LiveHeap, metrics.HeapGoal), metrics.HeapGoal > 0, m.style.Memory),
		m.metric("heap objects", formatCount(metrics.HeapObjects), m.style.Memory),
		m.metric("heap retained", formatBytes(metrics.HeapFree+metrics.HeapUnused), m.style.Memory),
		m.metric("heap released", formatBytes(metrics.HeapReleased), m.style.Memory),
		m.metric("runtime memory", formatBytes(metrics.RuntimeMemory), m.style.Memory),
		m.metric("stack memory", formatBytes(metrics.StackMemory), m.style.Memory),
		m.metricAvailable("rss gap", formatBytes(rssGap), process.Has(model.ProcessMemory), m.style.Memory),
		m.metric("memory limit", formatMemoryLimit(metrics.MemoryLimit), m.style.Memory),
		m.historyRow("live", heapHistory, m.style.Memory),
	}
}

func (m Model) nodeEventRows() []string {
	metrics := m.metrics.Node
	return []string{
		m.metric("utilization", formatPercent(metrics.EventLoopUtilization*100), m.style.CPU),
		m.metric("active resources", formatCount(metrics.ActiveResources), m.style.Value),
		m.metric("delay mean", formatDuration(metrics.EventLoopDelayMean), m.style.CPU),
		m.metric("delay p95", formatDuration(metrics.EventLoopDelayP95), m.style.CPU),
		m.metric("delay p99", formatDuration(metrics.EventLoopDelayP99), m.style.CPU),
		m.metric("delay max", formatDuration(metrics.EventLoopDelayMax), m.style.CPU),
	}
}

func (m Model) nodeMemoryRows() []string {
	metrics := m.metrics.Node
	process := m.metrics.Process
	return []string{
		m.metric("heap used", formatBytes(metrics.HeapUsed), m.style.Memory),
		m.metric("heap total", formatBytes(metrics.HeapTotal), m.style.Memory),
		m.metric("external", formatBytes(metrics.External), m.style.Memory),
		m.metric("array buffers", formatBytes(metrics.ArrayBuffers), m.style.Memory),
		m.metricAvailable("rss", formatBytes(process.RSS), process.Has(model.ProcessMemory), m.style.Memory),
		m.metricAvailable("virtual", formatBytes(process.VirtualMemory), process.Has(model.ProcessMemory), m.style.Memory),
		m.historyRow("heap", heapHistory, m.style.Memory),
	}
}

func (m Model) histogramRows(width, height int) []string {
	leftWidth := (width - 3) / 2
	rightWidth := width - leftWidth - 3
	return m.joinBlocks(
		m.histogramBlock("scheduler", m.scheduler.Histogram(), m.style.CPU, leftWidth, height),
		m.histogramBlock("gc pauses", m.gcPauses.Histogram(), m.style.Memory, rightWidth, height),
		leftWidth, rightWidth,
	)
}

func (m Model) histogramBlock(name string, histogram *model.Histogram, style lipgloss.Style, width, height int) []string {
	chartHeight := max(1, height-2)
	lines := []string{m.rowWidth(m.style.Label.Render(name), m.style.Label.Render("n "+formatCount(histogramCount(histogram))), width)}
	lines = append(lines, histogramChart(histogram, width, chartHeight, style)...)
	return append(lines, m.rowWidth(
		m.style.Label.Render("p95 "+formatQuantile(histogram, 0.95)),
		m.style.Label.Render("p99 "+formatQuantile(histogram, 0.99)), width,
	))
}

func (m Model) historyRows(width, height int) []string {
	first := max(1, (width-6)/3)
	last := max(1, width-2*first-6)
	left := m.historyBlock("cpu", cpuHistory, m.style.CPU, first, height)
	middle := m.historyBlock("rss", rssHistory, m.style.Memory, first, height)
	right := m.historyBlock("event loop p99", eventLoopHistory, m.style.CPU, last, height)
	return m.joinThreeBlocks(left, middle, right, first, last)
}

func (m Model) historyBlock(name string, kind historyKind, style lipgloss.Style, width, height int) []string {
	value, ok := latestHistoryValue(m.history, kind)
	lines := []string{m.rowWidth(m.style.Label.Render(name), style.Render(formatHistoryValue(value, kind, ok)), width)}
	return append(lines, historyChart(m.history, kind, width, max(1, height-1), style)...)
}

func (m Model) joinBlocks(left, right []string, leftWidth, rightWidth int) []string {
	lines := make([]string, max(len(left), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, leftWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(right, index, rightWidth)
	}
	return lines
}

func (m Model) joinThreeBlocks(left, middle, right []string, columnWidth, lastWidth int) []string {
	lines := make([]string, max(len(left), len(middle), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, columnWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(middle, index, columnWidth) + " " + m.style.Separator.Render("│") + " " +
			m.blockLine(right, index, lastWidth)
	}
	return lines
}

func (m Model) blockLine(lines []string, index, width int) string {
	if index >= len(lines) {
		return strings.Repeat(" ", max(0, width))
	}
	return m.style.Column.Width(width).Render(ansi.Truncate(lines[index], width, ""))
}

func (m Model) columns(left, right []string, width, columnWidth int) []string {
	lines := make([]string, max(len(left), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, columnWidth) + strings.Repeat(" ", 4) +
			m.blockLine(right, index, max(0, width-columnWidth-4))
	}
	return lines
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
	cpuHistory historyKind = iota
	heapHistory
	rssHistory
	eventLoopHistory
)

func (m Model) latestSample() sample {
	if len(m.history) == 0 {
		return sample{}
	}
	return m.history[len(m.history)-1]
}

func (m Model) historyRow(label string, kind historyKind, style lipgloss.Style) string {
	return m.metric(label, historySparkline(m.history, kind, 12), style)
}

func formatSeconds(seconds float64) string {
	return formatDuration(time.Duration(seconds * float64(time.Second)))
}

func formatDuration(duration time.Duration) string {
	if duration > 0 && duration < time.Millisecond {
		return duration.Round(time.Microsecond).String()
	}
	return duration.Round(time.Millisecond).String()
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

func formatPercent(value float64) string { return fmt.Sprintf("%.1f%%", value) }

func formatRatio(value, total uint64) string {
	if total == 0 {
		return "—"
	}
	return formatPercent(float64(value) / float64(total) * 100)
}

func formatByteRate(value float64) string { return formatBytes(uint64(max(0, value))) + "/s" }

func formatRate(value float64) string { return fmt.Sprintf("%.1f/s", value) }

func formatGOGC(value uint64) string {
	if value >= math.MaxInt64 {
		return "off"
	}
	return formatCount(value)
}

func formatMemoryLimit(value uint64) string {
	if value >= math.MaxInt64 {
		return "unlimited"
	}
	return formatBytes(value)
}

func histogramCount(histogram *model.Histogram) uint64 {
	if histogram == nil {
		return 0
	}
	var count uint64
	for _, bucket := range histogram.Counts {
		count += bucket
	}
	return count
}

func formatQuantile(histogram *model.Histogram, quantile float64) string {
	value, ok := histogramQuantile(histogram, quantile)
	if !ok {
		return "—"
	}
	return formatSeconds(value)
}

func histogramQuantile(histogram *model.Histogram, quantile float64) (float64, bool) {
	total := histogramCount(histogram)
	if total == 0 || histogram == nil || len(histogram.Buckets) != len(histogram.Counts)+1 {
		return 0, false
	}
	target := uint64(math.Ceil(float64(total) * quantile))
	var cumulative uint64
	for index, count := range histogram.Counts {
		cumulative += count
		if cumulative < target {
			continue
		}
		value := histogram.Buckets[index+1]
		if math.IsInf(value, 1) {
			value = histogram.Buckets[index]
		}
		return max(0, value), true
	}
	return 0, false
}

func histogramChart(histogram *model.Histogram, width, height int, style lipgloss.Style) []string {
	bins := histogramBins(histogram, width)
	lines := make([]string, max(0, height))
	var peak uint64
	for _, count := range bins {
		peak = max(peak, count)
	}
	for row := range lines {
		var line strings.Builder
		line.Grow(width * 3)
		threshold := uint64(height - row)
		for _, count := range bins {
			level := uint64(0)
			if peak > 0 {
				level = (count*uint64(height) + peak - 1) / peak
			}
			if level >= threshold {
				line.WriteRune('█')
			} else {
				line.WriteByte(' ')
			}
		}
		line.WriteString(strings.Repeat(" ", max(0, width-len(bins))))
		lines[row] = style.Render(line.String())
	}
	return lines
}

func histogramBins(histogram *model.Histogram, width int) []uint64 {
	if histogram == nil || len(histogram.Counts) == 0 || width <= 0 {
		return nil
	}
	first, last := 0, len(histogram.Counts)-1
	for first < last && histogram.Counts[first] == 0 {
		first++
	}
	for last > first && histogram.Counts[last] == 0 {
		last--
	}
	counts := histogram.Counts[first : last+1]
	bins := make([]uint64, width)
	if len(counts) == 1 {
		bins[width/2] = counts[0]
		return bins
	}
	for index, count := range counts {
		if count > 0 {
			bins[index*(width-1)/(len(counts)-1)] += count
		}
	}
	return bins
}

func historySparkline(history []sample, kind historyKind, width int) string {
	if len(history) == 0 || width <= 0 {
		return strings.Repeat("·", max(0, width))
	}
	start := max(0, len(history)-width)
	values := history[start:]
	peak := 0.0
	for _, sample := range values {
		if value, ok := historyValue(sample, kind); ok {
			peak = max(peak, value)
		}
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	var result strings.Builder
	result.Grow(width * 3)
	result.WriteString(strings.Repeat("·", width-len(values)))
	for _, sample := range values {
		value, ok := historyValue(sample, kind)
		if !ok {
			result.WriteRune('·')
			continue
		}
		level := 0
		if peak > 0 {
			level = int(value * float64(len(levels)-1) / peak)
		}
		result.WriteRune(levels[level])
	}
	return result.String()
}

func historyChart(history []sample, kind historyKind, width, height int, style lipgloss.Style) []string {
	start := max(0, len(history)-width)
	values := history[start:]
	peak := 0.0
	for _, sample := range values {
		if value, ok := historyValue(sample, kind); ok {
			peak = max(peak, value)
		}
	}
	lines := make([]string, max(0, height))
	for row := range lines {
		var line strings.Builder
		line.Grow(width * 3)
		line.WriteString(strings.Repeat(" ", max(0, width-len(values))))
		for _, sample := range values {
			value, ok := historyValue(sample, kind)
			level := 0
			if ok && peak > 0 {
				level = int(math.Ceil(value * float64(height) / peak))
			}
			if level >= height-row {
				line.WriteRune('█')
			} else {
				line.WriteByte(' ')
			}
		}
		lines[row] = style.Render(line.String())
	}
	return lines
}

func latestHistoryValue(history []sample, kind historyKind) (float64, bool) {
	for index := len(history) - 1; index >= 0; index-- {
		if value, ok := historyValue(history[index], kind); ok {
			return value, true
		}
	}
	return 0, false
}

func historyValue(sample sample, kind historyKind) (float64, bool) {
	switch kind {
	case cpuHistory:
		return sample.CPUCores, sample.available&sampleCPU != 0
	case heapHistory:
		return float64(sample.LiveMemory), !sample.At.IsZero()
	case rssHistory:
		return float64(sample.RSS), sample.available&sampleRSS != 0
	case eventLoopHistory:
		return float64(sample.EventLoopP99), !sample.At.IsZero()
	default:
		return 0, false
	}
}

func formatHistoryValue(value float64, kind historyKind, ok bool) string {
	if !ok {
		return "—"
	}
	switch kind {
	case cpuHistory:
		return fmt.Sprintf("%.2f", value)
	case heapHistory, rssHistory:
		return formatBytes(uint64(value))
	case eventLoopHistory:
		return formatDuration(time.Duration(value))
	default:
		return "—"
	}
}
