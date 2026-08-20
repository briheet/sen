package metrics

import (
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestMetricsUseTerminalBackground(t *testing.T) {
	panel := New(nil)
	background := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	panel, _ = panel.Update(tea.BackgroundColorMsg{Color: background})

	require.Equal(t, background, panel.style.Root.GetBackground())
	require.Equal(t, background, panel.style.Component.GetBackground())
}

func TestDashboardRendersRedisSnapshot(t *testing.T) {
	started := time.Unix(100, 0)
	panel := New(nil)
	panel.ApplySnapshot(redisSnapshot(100, 10, 1000, 2000), started)
	panel.ApplySnapshot(redisSnapshot(150, 14, 1400, 2600), started.Add(2*time.Second))
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 112, Height: 35})

	view := panel.View()
	plain := ansi.Strip(view)
	require.Equal(t, 112, lipgloss.Width(view))
	require.Equal(t, 35, lipgloss.Height(view))
	require.Contains(t, plain, "SERVER")
	require.Contains(t, plain, "MEMORY")
	require.Contains(t, plain, "ACTIVITY")
	require.Contains(t, plain, "HISTORY · 30S")
	require.Contains(t, plain, "8.0.0")
	require.Regexp(t, `clients\s+:\s+4`, plain)
	require.Regexp(t, `commands/s\s+:\s+25\.0/s`, plain)
	require.Regexp(t, `connections/s\s+:\s+2\.0/s`, plain)
	require.Regexp(t, `network in\s+:\s+200 B/s`, plain)
	require.Regexp(t, `cache hit\s+:\s+80\.0% window`, plain)
	require.Regexp(t, `cpu user / sys\s+:\s+500ms / 250ms`, plain)
}

func TestDashboardKeepsRoundedBorder(t *testing.T) {
	started := time.Unix(100, 0)
	panel := New(nil)
	panel.ApplySnapshot(redisSnapshot(100, 10, 1000, 2000), started)
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 112, Height: 35})

	lines := strings.Split(ansi.Strip(panel.View()), "\n")
	require.True(t, strings.HasPrefix(lines[0], "╭"))
	require.True(t, strings.HasSuffix(lines[0], "╮"))
	require.True(t, strings.HasPrefix(lines[len(lines)-1], "╰"))
	require.True(t, strings.HasSuffix(lines[len(lines)-1], "╯"))
	require.NotContains(t, strings.Join(lines, "\n"), "▏")
	require.NotContains(t, strings.Join(lines, "\n"), "▔")
}

func TestHistoryColumnsFillThirtySecondPlot(t *testing.T) {
	started := time.Unix(100, 0)
	history := make([]sample, 31)
	for second := range history {
		history[second] = sample{At: started.Add(time.Duration(second) * time.Second), Operations: float64(second + 1)}
	}

	columns := historyColumns(history, operationsHistory, 43)
	require.Len(t, columns, 43)
	for _, column := range columns {
		require.True(t, column.available)
	}
	require.Equal(t, 1.0, columns[0].value)
	require.Equal(t, 31.0, columns[len(columns)-1].value)
}

func TestHistoryChartShowsLatestZeroAtRightEdge(t *testing.T) {
	started := time.Unix(100, 0)
	history := []sample{
		{At: started.Add(-historyWindow), Operations: 1},
		{At: started, Operations: 0},
	}

	chart := historyChart(history, operationsHistory, 10, 4, lipgloss.NewStyle())
	require.Len(t, chart, 4)
	require.True(t, strings.HasSuffix(chart[len(chart)-1], "▁"))
}

func TestApplySnapshotExpiresOldHistory(t *testing.T) {
	panel := New(nil)
	started := time.Unix(100, 0)
	panel.ApplySnapshot(redisSnapshot(1, 1, 1, 1), started)
	panel.ApplySnapshot(redisSnapshot(2, 2, 2, 2), started.Add(31*time.Second))

	require.Len(t, panel.history, 1)
	require.Equal(t, started.Add(31*time.Second), panel.history[0].At)
}

func TestSizeUsesEightyPercent(t *testing.T) {
	width, height := Size(120, 50)
	require.Equal(t, 96, width)
	require.Equal(t, 40, height)
}

func redisSnapshot(commands, connections, networkInput, networkOutput uint64) model.RuntimeMetrics {
	return model.RuntimeMetrics{Redis: model.RedisMetrics{
		Version: "8.0.0", Mode: "standalone", Role: "master", Uptime: time.Minute,
		UsedMemory: 2 << 20, UsedMemoryDataset: 1 << 20, RSS: 3 << 20, PeakMemory: 4 << 20,
		MaxMemory: 8 << 20, MemoryFragmentationRatio: 1.5,
		UserCPU: 0.5, SystemCPU: 0.25,
		ConnectedClients: 4, BlockedClients: 1, Keys: 12, InstantaneousOps: 25,
		TotalCommandsProcessed: commands, TotalConnectionsReceived: connections,
		NetworkInputBytes: networkInput, NetworkOutputBytes: networkOutput,
		KeyspaceHits: commands * 4, KeyspaceMisses: commands,
		ExpiredKeys: 3, EvictedKeys: 1,
	}}
}
