package metrics

import (
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestMetricsUseTerminalBackground(t *testing.T) {
	panel := New(nil, config.ServiceLangGo)
	background := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	panel, _ = panel.Update(tea.BackgroundColorMsg{Color: background})

	require.Equal(t, background, panel.style.Root.GetBackground())
	require.Equal(t, background, panel.style.Component.GetBackground())
}

func TestMetricsPanelRendersGoSnapshot(t *testing.T) {
	metrics := model.RuntimeMetrics{
		Process: model.ProcessMetrics{
			UserCPU: 1.25, SystemCPU: 0.25, RSS: 8 << 20, PeakRSS: 9 << 20, VirtualMemory: 32 << 20,
			Threads: 6, OpenFiles: 12, Uptime: time.Minute,
			Available: model.ProcessCPU | model.ProcessMemory | model.ProcessThreads |
				model.ProcessOpenFiles | model.ProcessUptime,
		},
		Go: model.GoMetrics{
			UserCPU: 1.25, GCCPU: 0.05, GCCycles: 3,
			AllocatedBytes: 10 << 20, Allocations: 42, LiveHeap: 2 << 20,
			HeapObjects: 21, HeapGoal: 4 << 20, MemoryLimit: 64 << 20,
			RuntimeMemory: 8 << 20, StackMemory: 512 << 10, HeapReleased: 1 << 20,
			HeapFree: 256 << 10, HeapUnused: 256 << 10,
			GOGC: 100, Goroutines: 12, GOMAXPROCS: 8, MutexWait: 0.025,
		},
	}
	source := &model.RuntimeGraph{Global: model.Global{Process: metrics}}
	panel := New(source, config.ServiceLangGo)
	width, height := Size(80, 17)
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: width, Height: height})

	plain := ansi.Strip(panel.View())
	require.Equal(t, width, lipgloss.Width(panel.View()))
	require.Equal(t, height, lipgloss.Height(panel.View()))
	require.Contains(t, plain, "2.0 MiB")
	require.Regexp(t, `peak rss\s+:\s+9\.0 MiB`, plain)
	require.Regexp(t, `heap usage\s+:\s+50\.0%`, plain)
	require.Regexp(t, `heap objects\s+:\s+21`, plain)
	require.Regexp(t, `heap retained\s+:\s+512\.0 KiB`, plain)
	require.Regexp(t, `rss gap\s+:\s+1\.0 MiB`, plain)
	require.Regexp(t, `stack memory\s+:\s+512\.0 KiB`, plain)
	for range panel.maxOffset() {
		panel, _ = panel.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	plain = ansi.Strip(panel.View())
	require.Contains(t, plain, "1.25s")
	require.Regexp(t, `GOMAXPROCS\s+:\s+8`, plain)
	require.Regexp(t, `mutex wait\s+:\s+25ms`, plain)
	require.NotContains(t, plain, "M close")
}

func TestMetricsDashboardGroupsGoTelemetry(t *testing.T) {
	panel := New(nil, config.ServiceLangGo)
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 112, Height: 35})

	view := panel.View()
	plain := ansi.Strip(view)
	require.Equal(t, 112, lipgloss.Width(view))
	require.Equal(t, 35, lipgloss.Height(view))
	require.Contains(t, plain, "PROCESS")
	require.Contains(t, plain, "GO RUNTIME")
	require.Contains(t, plain, "MEMORY")
	require.Contains(t, plain, "HISTOGRAMS · 30S")
	require.Contains(t, plain, "│")
	require.Contains(t, plain, "─")
}

func TestNodeDashboardShowsRuntimeTelemetry(t *testing.T) {
	metrics := model.RuntimeMetrics{
		Process: model.ProcessMetrics{RSS: 8 << 20, VirtualMemory: 16 << 20, Available: model.ProcessMemory},
		Node: model.NodeMetrics{
			HeapUsed: 2 << 20, HeapTotal: 4 << 20, External: 512 << 10,
			EventLoopUtilization: 0.25, EventLoopDelayP99: 3 * time.Millisecond,
		},
	}
	panel := New(&model.RuntimeGraph{Global: model.Global{Process: metrics}}, config.ServiceLangNode)
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 112, Height: 35})

	plain := ansi.Strip(panel.View())
	require.Contains(t, plain, "EVENT LOOP")
	require.Contains(t, plain, "heap used")
	require.Contains(t, plain, "heap total")
	require.Contains(t, plain, "external")
	require.Contains(t, plain, "rss")
	require.Contains(t, plain, "25.0%")
	require.NotContains(t, plain, "allocations")
	require.NotContains(t, plain, "gc cycles")
}

func TestSummaryRowsAlignValues(t *testing.T) {
	panel := New(nil, config.ServiceLangGo)
	for _, rows := range [][]string{panel.processRows(), panel.goRuntimeRows(), panel.goMemoryRows()} {
		for _, row := range rows {
			plain := ansi.Strip(row)
			require.Equal(t, metricLabelWidth, strings.IndexRune(plain, ':'))
		}
	}
}

func TestHistogramChartFillsAllocatedArea(t *testing.T) {
	chart := histogramChart(
		&model.Histogram{Counts: []uint64{1, 4, 2, 8}, Buckets: []float64{0, 1, 2, 3, 4}},
		8, 4, lipgloss.NewStyle(),
	)
	require.Len(t, chart, 4)
	for _, line := range chart {
		require.Equal(t, 8, lipgloss.Width(line))
	}
	require.Contains(t, strings.Join(chart, "\n"), "█")
}

func TestHistogramBinsScaleActiveRange(t *testing.T) {
	require.Equal(t, []uint64{2, 0, 0, 0, 0, 0, 0, 0, 4},
		histogramBins(&model.Histogram{Counts: []uint64{0, 2, 0, 4, 0}}, 9))
	require.Equal(t, []uint64{0, 0, 3, 0, 0},
		histogramBins(&model.Histogram{Counts: []uint64{0, 3, 0}}, 5))
}

func TestApplySnapshotDerivesRatesAndExpiresHistory(t *testing.T) {
	panel := New(nil, config.ServiceLangGo)
	started := time.Unix(100, 0)
	available := model.ProcessCPU | model.ProcessIO | model.ProcessIOOperations
	panel.ApplySnapshot(model.RuntimeMetrics{
		Process: model.ProcessMetrics{UserCPU: 1, ReadBytes: 100, ReadOps: 10, Available: available},
		Go:      model.GoMetrics{AllocatedBytes: 100, Allocations: 10, GCAssist: 0.1, GCCycles: 1},
	}, started)
	panel.ApplySnapshot(model.RuntimeMetrics{
		Process: model.ProcessMetrics{UserCPU: 3, ReadBytes: 300, ReadOps: 30, Available: available},
		Go:      model.GoMetrics{AllocatedBytes: 300, Allocations: 30, GCAssist: 0.3, GCCycles: 3},
	}, started.Add(2*time.Second))

	latest := panel.latestSample()
	require.Equal(t, 1.0, latest.CPUCores)
	require.Equal(t, 100.0, latest.AllocationRate)
	require.Equal(t, 10.0, latest.AllocationOps)
	require.Equal(t, 1.0, latest.GCCycleRate)
	require.InDelta(t, 10.0, latest.GCAssist, 0.0001)
	require.Equal(t, 100.0, latest.ReadRate)
	require.Equal(t, 10.0, latest.ReadOps)

	panel.ApplySnapshot(model.RuntimeMetrics{}, started.Add(34*time.Second))
	require.Len(t, panel.history, 1)
	require.Equal(t, started.Add(34*time.Second), panel.history[0].At)
}

func TestHistogramWindowUsesOnlyRecentDeltas(t *testing.T) {
	var window histogramWindow
	started := time.Unix(100, 0)
	window.Add(started, &model.Histogram{Counts: []uint64{2, 1}, Buckets: []float64{0, 1, 2}})
	window.Add(started.Add(time.Second), &model.Histogram{Counts: []uint64{5, 2}, Buckets: []float64{0, 1, 2}})
	require.Equal(t, []uint64{3, 1}, window.Histogram().Counts)

	window.Add(started.Add(32*time.Second), &model.Histogram{Counts: []uint64{7, 4}, Buckets: []float64{0, 1, 2}})
	require.Equal(t, []uint64{2, 2}, window.Histogram().Counts)
}

func TestSizeUsesEightyPercent(t *testing.T) {
	width, height := Size(120, 50)
	require.Equal(t, 96, width)
	require.Equal(t, 40, height)
}
