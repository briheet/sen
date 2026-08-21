package db

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/adapters/postgres/analysis"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/stretchr/testify/require"
)

func TestDatabaseViewsSeparateStatementsAndTables(t *testing.T) {
	static := analysis.BuildGraph(
		[]analysis.Statement{{QueryID: 42, Query: "SELECT * FROM users"}},
		[]analysis.Table{{Name: "users"}},
	)
	statements, tables := databaseViews(model.BuildRuntimeGraph(analysis.ModulePath, static))

	require.Equal(t, map[string]bool{"postgres-server": true, "SELECT * FROM users": true}, graphNames(statements))
	require.Equal(t, map[string]bool{"postgres-server": true, "users": true}, graphNames(tables))
	require.Equal(t, static.Root, statements.Static.Root)
	require.Equal(t, static.Root, tables.Static.Root)
}

func graphNames(graph *model.RuntimeGraph) map[string]bool {
	names := make(map[string]bool, len(graph.Nodes))
	for _, node := range graph.Nodes {
		names[node.Static.Name] = true
	}
	return names
}

func TestPageRendersDatabaseGraphAndMetrics(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	graph := model.BuildRuntimeGraph(analysis.ModulePath, analysis.BuildGraph(
		[]analysis.Statement{{QueryID: 42, Query: "SELECT * FROM users"}},
		[]analysis.Table{{Name: "users"}},
	))
	graph.ApplyUpdate(graph.BuildUpdate(&model.RuntimeMetrics{Postgres: model.PostgresMetrics{
		Version: "17.6", Database: "sen", DatabaseSize: 2 << 20,
	}}, nil, nil))
	page := New(&engine.Engine{
		Service: config.Service{Name: "database", Type: config.ServiceTypeDB, Provider: config.ServiceProviderPostgres},
		Graph:   graph,
	}, nil)

	updated, _ := page.Update(pages.ViewportMsg{Width: 80, Height: 18, Visible: true})
	page = updated.(Model)
	require.Equal(t, 80, lipgloss.Width(page.View().Content))
	require.Contains(t, page.View().Content, "Pixel graph requires")

	updated, _ = page.Update(tea.KeyPressMsg{Code: 'm', Mod: tea.ModShift, Text: "M"})
	page = updated.(Model)
	require.Contains(t, page.View().Content, "DATABASE")
	require.Contains(t, page.View().Content, "2.0 MiB")
}

func TestActivityAddsOnlySyntheticDatabaseEdge(t *testing.T) {
	static := analysis.BuildGraph(
		[]analysis.Statement{{QueryID: 42, Query: "SELECT 1"}},
		[]analysis.Table{{Name: "users"}},
	)
	statement := static.Nodes[static.Root].Out[0]
	snapshot := model.RuntimeSnapshot{
		NodeActivity: map[model.NodeID]int64{statement: 3},
		NodeEdges:    make(map[model.NodeEdge]int64),
	}

	addActivityEdges(static, &snapshot)
	require.Equal(t, int64(3), snapshot.NodeEdges[model.NodeEdge{From: static.Root, To: statement}])
}

func TestPageSwitchesViewsFromShiftKeybindings(t *testing.T) {
	static := analysis.BuildGraph(
		[]analysis.Statement{{QueryID: 42, Query: "SELECT * FROM users"}},
		[]analysis.Table{{Name: "users"}},
	)
	graph := model.BuildRuntimeGraph(analysis.ModulePath, static)
	page := New(&engine.Engine{
		Service: config.Service{Name: "database", Type: config.ServiceTypeDB, Provider: config.ServiceProviderPostgres},
		Graph:   graph,
	}, nil)

	updated, _ := page.Update(pages.ViewportMsg{Width: 80, Height: 18, Visible: true})
	page = updated.(Model)

	updated, command := page.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModShift, Text: "L"})
	page = updated.(Model)
	require.Equal(t, 0, page.pager.Page)
	require.Equal(t, 1, page.pending)
	require.NotNil(t, command)

	updated, command = page.Update(switchGraphMsg{service: "database", page: 1})
	page = updated.(Model)
	require.Equal(t, 1, page.pager.Page)
	require.NotNil(t, command)

	updated, command = page.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModShift, Text: "H"})
	page = updated.(Model)
	require.Equal(t, 1, page.pager.Page)
	require.Equal(t, 0, page.pending)
	require.NotNil(t, command)

	updated, _ = page.Update(switchGraphMsg{service: "database", page: 0})
	page = updated.(Model)
	require.Equal(t, 0, page.pager.Page)
}
