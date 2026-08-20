package postgres

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

func TestDashboardRendersPostgresSnapshotAndRates(t *testing.T) {
	started := time.Unix(100, 0)
	panel := New(nil)
	panel.ApplySnapshot(postgresSnapshot(100, 10, 200, 20), started)
	panel.ApplySnapshot(postgresSnapshot(140, 12, 260, 24), started.Add(2*time.Second))
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 112, Height: 35})

	view := panel.View()
	plain := ansi.Strip(view)
	require.Equal(t, 112, lipgloss.Width(view))
	require.Equal(t, 35, lipgloss.Height(view))
	require.Contains(t, plain, "DATABASE")
	require.Contains(t, plain, "ACTIVITY")
	require.Contains(t, plain, "STORAGE")
	require.Contains(t, plain, "HISTORY · 30S")
	require.Regexp(t, `transactions/s\s+:\s+21\.0/s`, plain)
	require.Regexp(t, `queries/s\s+:\s+30\.0/s`, plain)
	require.Regexp(t, `connections\s+:\s+4 / 100`, plain)
}

func TestDashboardKeepsRoundedBorder(t *testing.T) {
	panel := New(nil)
	panel.ApplySnapshot(postgresSnapshot(1, 0, 1, 1), time.Unix(100, 0))
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 112, Height: 35})
	lines := strings.Split(ansi.Strip(panel.View()), "\n")
	require.True(t, strings.HasPrefix(lines[0], "╭"))
	require.True(t, strings.HasSuffix(lines[0], "╮"))
	require.True(t, strings.HasPrefix(lines[len(lines)-1], "╰"))
	require.True(t, strings.HasSuffix(lines[len(lines)-1], "╯"))
}

func TestDashboardUsesTerminalBackground(t *testing.T) {
	panel := New(nil)
	background := color.RGBA{R: 12, G: 13, B: 14, A: 255}
	panel, _ = panel.Update(tea.BackgroundColorMsg{Color: background})
	require.Equal(t, background, panel.style.Root.GetBackground())
	require.Equal(t, background, panel.style.Component.GetBackground())
}

func postgresSnapshot(commits, rollbacks, statementCalls, blocksRead uint64) model.RuntimeMetrics {
	return model.RuntimeMetrics{Postgres: model.PostgresMetrics{
		Version: "17.6", Database: "sen", Uptime: time.Hour, DatabaseSize: 8 << 20,
		MaxConnections: 100, Backends: 4, Active: 2, Idle: 2, Locks: 3,
		Commits: commits, Rollbacks: rollbacks, StatementCalls: statementCalls, StatementsAvailable: true,
		BlocksRead: blocksRead, BlocksHit: blocksRead * 10,
		TuplesReturned: commits * 2, TuplesFetched: commits,
		TempBytes: blocksRead * 1024,
	}}
}
