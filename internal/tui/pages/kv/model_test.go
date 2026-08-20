package kv

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/adapters/redis/analysis"
	redistrace "github.com/briheet/sen/internal/adapters/redis/runtime/trace"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/stretchr/testify/require"
)

type observationRuntime struct{ observation model.Observation }

func (observationRuntime) Start(context.Context) error { return nil }
func (r observationRuntime) Collect(context.Context) (model.Observation, error) {
	return r.observation, nil
}
func (observationRuntime) Wait() error    { return nil }
func (observationRuntime) Stop() error    { return nil }
func (observationRuntime) Cleanup() error { return nil }

func TestPageRendersCommandGraphAndMetricsOverlay(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	runtimeGraph := model.BuildRuntimeGraph(analysis.ModulePath, analysis.BuildGraph())
	runtimeGraph.ApplyUpdate(runtimeGraph.BuildUpdate(&model.RuntimeMetrics{Redis: model.RedisMetrics{
		Version: "8.0.0", UsedMemory: 2 << 20, InstantaneousOps: 25,
	}}, nil, nil))
	page := New(&engine.Engine{
		Service: config.Service{
			Name: "cache", Type: config.ServiceTypeKV,
			Provider: config.ServiceProviderRedis, Address: "localhost:6379",
		},
		Graph: runtimeGraph,
	}, nil)

	updated, _ := page.Update(pages.ViewportMsg{Width: 80, Height: 18, Visible: true})
	page = updated.(Model)
	require.Equal(t, 80, lipgloss.Width(page.View().Content))
	require.Equal(t, 18, lipgloss.Height(page.View().Content))
	require.Contains(t, page.View().Content, "Pixel graph requires")

	updated, _ = page.Update(tea.KeyPressMsg{Code: 'm', Mod: tea.ModShift, Text: "M"})
	page = updated.(Model)
	require.Contains(t, page.View().Content, "commands/s")
	require.Contains(t, page.View().Content, "2.0 MiB")

	updated, _ = page.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	page = updated.(Model)
	require.NotContains(t, page.View().Content, "commands/s")
}

func TestPlaceholderPageReportsUnavailableTelemetry(t *testing.T) {
	page := FromService(config.Service{
		Name: "cache", Type: config.ServiceTypeKV,
		Provider: config.ServiceProviderRedis, Address: "localhost:6379",
	})
	updated, _ := page.Update(pages.ViewportMsg{Width: 60, Height: 12, Visible: true})

	view := updated.View()
	require.Contains(t, view.Content, "cache (redis)")
	require.Contains(t, view.Content, "Telemetry unavailable")
}

func TestCommandActivityActivatesOnlyItsSyntheticEdge(t *testing.T) {
	static := analysis.BuildGraph()
	get, set := commandNode(t, static, "GET"), commandNode(t, static, "SET")
	snapshot := model.RuntimeSnapshot{
		NodeActivity: map[model.NodeID]int64{get: 3},
		NodeEdges:    make(map[model.NodeEdge]int64),
	}

	addCommandEdges(static, &snapshot)

	require.Equal(t, int64(3), snapshot.NodeEdges[model.NodeEdge{From: static.Root, To: get}])
	require.NotContains(t, snapshot.NodeEdges, model.NodeEdge{From: static.Root, To: set})
}

func TestRedisProfileReachesCommandEdges(t *testing.T) {
	t.Setenv("TERM", "dumb")
	runtimeGraph := model.BuildRuntimeGraph(analysis.ModulePath, analysis.BuildGraph())
	target := &engine.Engine{
		Service: config.Service{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis},
		Runtime: observationRuntime{observation: model.Observation{
			Metrics: &model.RuntimeMetrics{},
			Profiles: map[string]*model.Profile{
				redistrace.ProfileName: redistrace.Parse("cmdstat_get:calls=3,usec=9\r\n").Profile(time.Second),
			},
		}},
		Graph: runtimeGraph,
	}
	require.NoError(t, target.Refresh(context.Background()))

	var debug bytes.Buffer
	page := New(target, &debug)
	_, _ = page.Update(pages.TelemetryTickMsg{})

	require.Contains(t, debug.String(), "active_nodes=1 active_edges=1")
}

func commandNode(t *testing.T, static *model.StaticGraph, name string) model.NodeID {
	t.Helper()
	for id, node := range static.Nodes {
		if node.Name == name {
			return id
		}
	}
	t.Fatalf("command node %q not found", name)
	return 0
}
