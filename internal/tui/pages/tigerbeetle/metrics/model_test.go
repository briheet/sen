package metrics

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/require"
)

func TestApplySnapshotAndView(t *testing.T) {
	now := time.Now()
	panel := New(nil, 2)
	panel.ApplySnapshot(model.RuntimeMetrics{TigerBeetle: model.TigerBeetleMetrics{
		Cluster: "0123456789abcdef0123456789abcdef", Window: 10 * time.Second,
		Replicas: map[uint32]model.TigerBeetleReplicaMetrics{
			0: {ObservedAt: now, Status: 0, View: 7, GridCacheHits: 90, GridCacheMisses: 10},
			1: {ObservedAt: now, Status: 0, View: 7, GridCacheHits: 90, GridCacheMisses: 10},
		},
		Operations: map[string]model.TigerBeetleOperationMetrics{
			"create_accounts": {Requests: 20, LatencySum: 400 * time.Microsecond, LatencyMax: 40 * time.Microsecond},
		},
	}}, now)
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 120, Height: 32})
	view := panel.View()
	require.Contains(t, view, "healthy")
	require.Contains(t, view, "2.0/s")
	require.Contains(t, view, "90.0%")
	require.Equal(t, float64(2), panel.history[0].Requests)
}

func TestStaleReplicaIsDegraded(t *testing.T) {
	now := time.Now()
	panel := New(nil, 1)
	panel.ApplySnapshot(model.RuntimeMetrics{TigerBeetle: model.TigerBeetleMetrics{
		Cluster: "0123456789abcdef0123456789abcdef", Window: 10 * time.Second,
		Replicas:   map[uint32]model.TigerBeetleReplicaMetrics{0: {ObservedAt: now.Add(-30 * time.Second)}},
		Operations: map[string]model.TigerBeetleOperationMetrics{},
	}}, now)
	panel, _ = panel.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, panel.View(), "degraded")
}
