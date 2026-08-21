package tui

import (
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	redisanalysis "github.com/briheet/sen/internal/adapters/redis/analysis"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	runtimeModel "github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages/db"
	"github.com/briheet/sen/internal/tui/pages/kv"
	"github.com/briheet/sen/internal/tui/pages/servers"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/stretchr/testify/require"
)

func TestModelContainsBuiltEngines(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	engines := []*engine.Engine{
		{Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo}},
		{Service: config.Service{Name: "worker", Type: config.ServiceTypeServer, Lang: config.ServiceLangNode}},
	}
	services := []config.Service{
		engines[0].Service,
		engines[1].Service,
		{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis, Address: "localhost:6379"},
		{Name: "database", Type: config.ServiceTypeDB, Provider: config.ServiceProviderPostgres, Address: "postgres://localhost/sen"},
	}
	m := initialModel(engines, services, "/cache/sen/project/engine.log", nil)
	require.Len(t, m.ctx.Pages(), 4)
	apiPage, _ := m.ctx.Page("api")
	workerPage, _ := m.ctx.Page("worker")
	cachePage, _ := m.ctx.Page("cache")
	databasePage, _ := m.ctx.Page("database")
	api := apiPage.(servers.Model)
	worker := workerPage.(servers.Model)
	cache := cachePage.(kv.Model)
	database := databasePage.(db.Model)
	require.Same(t, engines[0], api.Engine)
	require.Same(t, engines[1], worker.Engine)
	require.Equal(t, "localhost:6379", cache.Service.Address)
	require.Nil(t, cache.Engine)
	require.Nil(t, database.Engine)
	require.Equal(t, "postgres://localhost/sen", database.Service.Address)
	require.Equal(t, "api", m.ctx.ActivePage())
	require.Equal(t, styles.Zakura, m.activeTheme)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updatedModel := updated.(model)
	view := updatedModel.View()
	require.Contains(t, view.Content, "api")
	require.Contains(t, view.Content, "sen")
	require.Equal(t, 80, lipgloss.Width(view.Content))
	require.Equal(t, 24, lipgloss.Height(view.Content))
	require.True(t, view.AltScreen)
	require.Equal(t, tea.MouseModeCellMotion, view.MouseMode)
	require.Equal(t, 1, lipgloss.Height(updatedModel.statusbar.View()))
	require.NotContains(t, view.Content, "╭")
}

func TestModelUsesConfiguredTheme(t *testing.T) {
	m := initialModel(nil, nil, "", nil, styles.CatppuccinMocha)

	require.Equal(t, styles.CatppuccinMocha, m.activeTheme)
}

func TestGraphRendersAboveStatusBar(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	static := &runtimeModel.StaticGraph{
		Root:  1,
		Nodes: map[runtimeModel.NodeID]*runtimeModel.StaticNode{1: {ID: 1, Name: "main"}},
	}
	target := &engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph: &runtimeModel.RuntimeGraph{
			Static: static,
			Nodes:  map[runtimeModel.NodeID]*runtimeModel.Node{1: {Static: static.Nodes[1]}},
		},
	}
	m := initialModel([]*engine.Engine{target}, []config.Service{target.Service}, "", nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	lines := strings.Split(updated.(model).View().Content, "\n")
	require.Contains(t, lines[len(lines)-1], "api")
	require.Contains(t, updated.(model).View().Content, "main")
}

func TestModelMapsKVEngine(t *testing.T) {
	target := &engine.Engine{Service: config.Service{
		Name:     "cache",
		Type:     config.ServiceTypeKV,
		Provider: config.ServiceProviderRedis,
	}}
	model := initialModel([]*engine.Engine{target}, []config.Service{target.Service}, "", nil)

	cachePage, _ := model.ctx.Page("cache")
	cache := cachePage.(kv.Model)
	require.Same(t, target, cache.Engine)
	require.Equal(t, "cache", model.ctx.ActivePage())
}

func TestModelDumpsMessages(t *testing.T) {
	var dump bytes.Buffer
	m := initialModel(nil, nil, "", &dump)

	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	require.Contains(t, dump.String(), "(tea.WindowSizeMsg)")
	require.Contains(t, dump.String(), "Width: (int) 80")
}

func TestModelOmitsRawImagePayloadFromDump(t *testing.T) {
	var dump bytes.Buffer
	m := initialModel(nil, nil, "", &dump)

	_, _ = m.Update(tea.RawMsg{Msg: "encoded-image"})

	require.Contains(t, dump.String(), "image payload omitted")
	require.NotContains(t, dump.String(), "encoded-image")
}

func TestModelViewUsesCachedTerminalText(t *testing.T) {
	m := initialModel(nil, nil, "", nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(model)

	require.Zero(t, testing.AllocsPerRun(100, func() { _ = m.View() }))
}

func BenchmarkRefreshView(b *testing.B) {
	b.Setenv("TERM", "xterm-ghostty")
	target := &engine.Engine{
		Service: config.Service{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis},
		Graph:   runtimeModel.BuildRuntimeGraph(redisanalysis.ModulePath, redisanalysis.BuildGraph()),
	}
	m := initialModel([]*engine.Engine{target}, []config.Service{target.Service}, "", nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 173, Height: 45})
	m = updated.(model)
	b.ReportAllocs()

	for b.Loop() {
		m.refreshView()
	}
}
