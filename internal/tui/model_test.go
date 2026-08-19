package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages/kv"
	"github.com/briheet/sen/internal/tui/pages/servers"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/stretchr/testify/require"
)

func TestModelContainsBuiltEngines(t *testing.T) {
	engines := []*engine.Engine{
		{Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo}},
		{Service: config.Service{Name: "worker", Type: config.ServiceTypeServer, Lang: config.ServiceLangNode}},
	}
	services := []config.Service{
		engines[0].Service,
		engines[1].Service,
		{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis, Address: "localhost:6379"},
	}
	m := initialModel(engines, services, "/cache/sen/project/engine.log")
	require.Len(t, m.ctx.Pages(), 3)
	apiPage, _ := m.ctx.Page("api")
	workerPage, _ := m.ctx.Page("worker")
	cachePage, _ := m.ctx.Page("cache")
	api := apiPage.(servers.Model)
	worker := workerPage.(servers.Model)
	cache := cachePage.(kv.Model)
	require.Same(t, engines[0], api.Engine)
	require.Same(t, engines[1], worker.Engine)
	require.Equal(t, "localhost:6379", cache.Service.Address)
	require.Nil(t, cache.Engine)
	require.Equal(t, "api", m.ctx.ActivePage())
	require.Equal(t, styles.Zakura, m.activeTheme)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := updated.(model).View()
	require.Contains(t, view.Content, "api (go)")
	require.Contains(t, view.Content, "sen 0.1.0")
	require.Equal(t, 80, lipgloss.Width(view.Content))
	require.Equal(t, 24, lipgloss.Height(view.Content))
	require.True(t, view.AltScreen)
}

func TestModelMapsKVEngine(t *testing.T) {
	target := &engine.Engine{Service: config.Service{
		Name:     "cache",
		Type:     config.ServiceTypeKV,
		Provider: config.ServiceProviderRedis,
	}}
	model := initialModel([]*engine.Engine{target}, []config.Service{target.Service}, "")

	cachePage, _ := model.ctx.Page("cache")
	cache := cachePage.(kv.Model)
	require.Same(t, target, cache.Engine)
	require.Equal(t, "cache", model.ctx.ActivePage())
	require.Contains(t, model.View().Content, "cache (redis)")
}
