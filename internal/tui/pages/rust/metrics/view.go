package metrics

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
)

const labelWidth = 15

// View renders a responsive Rust dashboard.
func (m Model) View() string {
	if m.width < MinimumWidth || m.height < MinimumHeight {
		return ""
	}
	if m.width >= 96 && m.height >= 28 {
		return m.dashboard()
	}
	return m.compact()
}

func (m Model) dashboard() string {
	width, height := max(0, m.width-2), max(0, m.height-2)
	summaryHeight := max(10, (height-1)*48/100)
	detailHeight := max(0, height-summaryHeight-1)
	column := max(0, (width-2)/3)
	last := max(0, width-column*2-2)
	divider := m.style.Separator.Render(strings.TrimSuffix(strings.Repeat("│\n", summaryHeight), "\n"))
	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel("PROCESS", m.processRows(), column, summaryHeight), divider,
		m.panel("NATIVE PROFILE", m.profileRows(), column, summaryHeight), divider,
		m.panel("TOKIO RUNTIME", m.tokioRows(), last, summaryHeight),
	)
	history := m.panel("HISTORY · 30S", m.historyRows(width-2, detailHeight-1), width, detailHeight)
	body := lipgloss.JoinVertical(lipgloss.Left, summary, m.style.Separator.Render(strings.Repeat("─", width)), history)
	return m.style.Root.Width(m.width).Height(m.height).Render(body)
}

func (m Model) compact() string {
	width := max(0, m.width-2)
	column := max(0, (width-4)/2)
	left := append([]string{m.style.Section.Render("PROCESS")}, m.processRows()...)
	left = append(left, "", m.style.Section.Render("NATIVE PROFILE"))
	left = append(left, m.profileRows()...)
	right := append([]string{m.style.Section.Render("TOKIO RUNTIME")}, m.tokioRows()...)
	right = append(right, "", m.style.Section.Render("HISTORY · 30S"),
		m.metric("cpu", m.sparkline(cpuHistory, 14), m.style.CPU),
		m.metric("rss", m.sparkline(rssHistory, 14), m.style.Memory),
		m.metric("live tasks", m.sparkline(tasksHistory, 14), m.style.Tokio),
	)
	lines := make([]string, max(len(left), len(right)))
	for index := range lines {
		lines[index] = m.blockLine(left, index, column) + strings.Repeat(" ", 4) + m.blockLine(right, index, max(0, width-column-4))
	}
	visible := max(0, m.height-2)
	start := min(m.offset, max(0, len(lines)-visible))
	return m.style.Root.Width(m.width).Height(m.height).Render(strings.Join(lines[start:min(len(lines), start+visible)], "\n"))
}

func (m Model) processRows() []string {
	p := m.metrics.Process
	latest := m.latest()
	return []string{
		m.available("uptime", formatDuration(p.Uptime), p.Has(model.ProcessUptime), m.style.Value),
		m.available("cpu", fmt.Sprintf("%.2f cores", latest.cpu), latest.hasCPU, m.style.CPU),
		m.available("cpu time u/s", formatSeconds(p.UserCPU)+" / "+formatSeconds(p.SystemCPU), p.Has(model.ProcessCPU), m.style.CPU),
		m.available("threads", count(p.Threads), p.Has(model.ProcessThreads), m.style.Value),
		m.available("rss", bytes(p.RSS), p.Has(model.ProcessMemory), m.style.Memory),
		m.available("peak rss", bytes(p.PeakRSS), p.Has(model.ProcessMemory), m.style.Memory),
		m.available("open files", count(p.OpenFiles), p.Has(model.ProcessOpenFiles), m.style.Value),
	}
}

func (m Model) profileRows() []string {
	r := m.metrics.Rust
	return []string{
		m.metric("samples", count(r.ProfileSamples), m.style.CPU),
		m.metric("unique stacks", count(r.ProfileStacks), m.style.Value),
		m.metric("window", "1s native", m.style.Value),
		m.metric("frame pointers", "enabled", m.style.Value),
		m.metric("symbols", "DWARF", m.style.Value),
	}
}

func (m Model) tokioRows() []string {
	r := m.metrics.Rust
	if !r.TokioEnabled {
		return []string{
			m.metric("console", "off", m.style.Label),
			m.metric("enable", "tokio_console", m.style.Label),
		}
	}
	return []string{
		m.metric("console", "connected", m.style.Tokio),
		m.metric("tasks live", count(r.LiveTasks), m.style.Tokio),
		m.metric("tasks total", count(r.TotalTasks), m.style.Value),
		m.metric("completed", count(r.CompletedTasks), m.style.Value),
		m.metric("polls", count(r.Polls), m.style.CPU),
		m.metric("wakes / self", count(r.Wakes)+" / "+count(r.SelfWakes), m.style.Value),
		m.metric("busy time", formatDuration(r.BusyTime), m.style.CPU),
		m.metric("scheduled", formatDuration(r.ScheduledTime), m.style.Value),
		m.metric("resources", count(r.LiveResources), m.style.Memory),
		m.metric("async ops", count(r.LiveAsyncOperations), m.style.Memory),
		m.metric("dropped", count(r.DroppedEvents), m.style.Value),
	}
}

type historyKind uint8

const (
	cpuHistory historyKind = iota
	rssHistory
	tasksHistory
)

func (m Model) historyRows(width, height int) []string {
	column := max(1, (width-6)/3)
	last := max(1, width-column*2-6)
	blocks := [][]string{
		m.historyBlock("cpu", cpuHistory, column, height, m.style.CPU),
		m.historyBlock("rss", rssHistory, column, height, m.style.Memory),
		m.historyBlock("live tasks", tasksHistory, last, height, m.style.Tokio),
	}
	lines := make([]string, max(len(blocks[0]), len(blocks[1]), len(blocks[2])))
	for i := range lines {
		lines[i] = m.blockLine(blocks[0], i, column) + " │ " + m.blockLine(blocks[1], i, column) + " │ " + m.blockLine(blocks[2], i, last)
	}
	return lines
}

func (m Model) historyBlock(name string, kind historyKind, width, height int, style lipgloss.Style) []string {
	value, available := m.latestValue(kind)
	header := m.style.Label.Render(name) + strings.Repeat(" ", max(1, width-len(name)-len(formatValue(value, kind, available)))) + style.Render(formatValue(value, kind, available))
	return append([]string{header}, chart(m.values(kind), width, max(1, height-1), style)...)
}

func chart(values []float64, width, height int, style lipgloss.Style) []string {
	values = tail(values, width)
	maximum := 0.0
	for _, value := range values {
		maximum = math.Max(maximum, value)
	}
	if maximum == 0 {
		maximum = 1
	}
	lines := make([]string, height)
	for row := 0; row < height; row++ {
		var line strings.Builder
		line.WriteString(strings.Repeat(" ", max(0, width-len(values))))
		threshold := float64(height-row-1) / float64(height)
		for _, value := range values {
			if value/maximum > threshold {
				line.WriteRune('█')
			} else {
				line.WriteByte(' ')
			}
		}
		lines[row] = style.Render(line.String())
	}
	return lines
}

func (m Model) sparkline(kind historyKind, width int) string {
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	values := tail(m.values(kind), width)
	maximum := 0.0
	for _, value := range values {
		maximum = math.Max(maximum, value)
	}
	var result strings.Builder
	result.WriteString(strings.Repeat(" ", max(0, width-len(values))))
	for _, value := range values {
		level := 0
		if maximum > 0 {
			level = int(math.Round(value / maximum * 8))
		}
		result.WriteRune(blocks[min(8, max(0, level))])
	}
	return result.String()
}

func (m Model) values(kind historyKind) []float64 {
	values := make([]float64, 0, len(m.history))
	for _, sample := range m.history {
		switch kind {
		case cpuHistory:
			if sample.hasCPU {
				values = append(values, sample.cpu)
			}
		case rssHistory:
			if sample.hasRSS {
				values = append(values, float64(sample.rss))
			}
		case tasksHistory:
			if sample.hasTokio {
				values = append(values, float64(sample.liveTasks))
			}
		}
	}
	return values
}

func (m Model) latestValue(kind historyKind) (float64, bool) {
	values := m.values(kind)
	if len(values) == 0 {
		return 0, false
	}
	return values[len(values)-1], true
}

func formatValue(value float64, kind historyKind, available bool) string {
	if !available {
		return "—"
	}
	switch kind {
	case cpuHistory:
		return fmt.Sprintf("%.2f", value)
	case rssHistory:
		return bytes(uint64(value))
	default:
		return count(uint64(value))
	}
}

func (m Model) panel(title string, rows []string, width, height int) string {
	rows = rows[:min(len(rows), max(0, height-1))]
	return m.style.Component.Width(width).Height(height).Render(strings.Join(append([]string{m.style.Section.Render(title)}, rows...), "\n"))
}

func (m Model) blockLine(lines []string, index, width int) string {
	if index >= len(lines) {
		return strings.Repeat(" ", width)
	}
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(lines[index])
}

func (m Model) metric(label, value string, style lipgloss.Style) string {
	label += strings.Repeat(" ", max(0, labelWidth-lipgloss.Width(label))) + ":"
	return m.style.Label.Render(label) + " " + style.Render(value)
}

func (m Model) available(label, value string, ok bool, style lipgloss.Style) string {
	if !ok {
		value = "—"
	}
	return m.metric(label, value, style)
}

func (m Model) latest() sample {
	if len(m.history) == 0 {
		return sample{}
	}
	return m.history[len(m.history)-1]
}
func tail(values []float64, width int) []float64 {
	if len(values) > width {
		return values[len(values)-width:]
	}
	return values
}
func count(value uint64) string { return strconv.FormatUint(value, 10) }
func formatSeconds(value float64) string {
	return formatDuration(time.Duration(value * float64(time.Second)))
}
func formatDuration(value time.Duration) string {
	if value == 0 {
		return "0s"
	}
	return value.Round(time.Millisecond).String()
}
func bytes(value uint64) string {
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	amount := float64(value)
	unit := "B"
	for _, next := range units {
		amount /= 1024
		unit = next
		if amount < 1024 {
			break
		}
	}
	return fmt.Sprintf("%.1f %s", amount, unit)
}
