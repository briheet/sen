package tigerbeetle

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/adapters/tigerbeetle/analysis"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/stretchr/testify/require"
)

func TestViewsSeparateOperationsAndReplicas(t *testing.T) {
	static := analysis.BuildGraph([]string{"127.0.0.1:3000", "127.0.0.1:3001"})
	operations, replicas := tigerBeetleViews(model.BuildRuntimeGraph(analysis.ModulePath, static))
	require.Len(t, operations.Nodes, 1+len(analysis.Operations))
	require.Len(t, replicas.Nodes, 3)
	require.Contains(t, graphNames(operations), "create_accounts")
	require.Contains(t, graphNames(replicas), "replica 1")
	require.NotContains(t, graphNames(replicas), "create_accounts")
}

func TestPageRendersGraphAndMetrics(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	graph := model.BuildRuntimeGraph(analysis.ModulePath, analysis.BuildGraph([]string{"127.0.0.1:3000"}))
	graph.ApplyUpdate(graph.BuildUpdate(&model.RuntimeMetrics{TigerBeetle: model.TigerBeetleMetrics{
		Cluster: "0123456789abcdef0123456789abcdef", Window: 10 * time.Second,
		Replicas: map[uint32]model.TigerBeetleReplicaMetrics{0: {ObservedAt: time.Now(), View: 9}},
		Operations: map[string]model.TigerBeetleOperationMetrics{
			"create_accounts": {Requests: 10, LatencySum: 200 * time.Microsecond, LatencyMax: 30 * time.Microsecond},
		},
	}}, nil, nil))
	page := New(&engine.Engine{Service: config.Service{
		Name: "ledger", Type: config.ServiceTypeDB, Provider: config.ServiceProviderTigerBeetle,
		Addresses: []string{"127.0.0.1:3000"}, MetricsAddress: "127.0.0.1:8125",
	}, Graph: graph}, nil)

	updated, _ := page.Update(pages.ViewportMsg{Width: 100, Height: 30, Visible: true})
	page = updated.(Model)
	require.Equal(t, 100, lipgloss.Width(page.View().Content))
	require.Contains(t, page.View().Content, "Pixel graph requires")
	updated, _ = page.Update(tea.KeyPressMsg{Code: 'm', Mod: tea.ModShift, Text: "M"})
	page = updated.(Model)
	require.Contains(t, page.View().Content, "CLUSTER")
	require.Contains(t, page.View().Content, "requests/s")
}

func TestActivityAddsSyntheticEdge(t *testing.T) {
	static := analysis.BuildGraph([]string{"127.0.0.1:3000"})
	active := static.Nodes[static.Root].Out[0]
	snapshot := model.RuntimeSnapshot{NodeActivity: map[model.NodeID]int64{active: 1}, NodeEdges: make(map[model.NodeEdge]int64)}
	addActivityEdges(static, &snapshot)
	require.Equal(t, int64(1), snapshot.NodeEdges[model.NodeEdge{From: static.Root, To: active}])
}

func graphNames(graph *model.RuntimeGraph) map[string]bool {
	result := make(map[string]bool, len(graph.Nodes))
	for _, node := range graph.Nodes {
		result[node.Static.Name] = true
	}
	return result
}
