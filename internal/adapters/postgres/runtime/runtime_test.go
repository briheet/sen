package runtime

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/briheet/sen/internal/adapters/postgres/analysis"
	"github.com/briheet/sen/internal/adapters/postgres/runtime/trace"
	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUser = "postgres"
	testPort = "5433"
)

// testDSN returns a connection string if a server is reachable, else "".
func testDSN(t *testing.T) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+testPort, 200*time.Millisecond)
	if err != nil {
		return ""
	}
	_ = conn.Close()
	return "postgres://" + testUser + "@127.0.0.1:" + testPort + "/postgres"
}

func TestCollectorLive(t *testing.T) {
	dsn := testDSN(t)
	if dsn == "" {
		t.Skip("no postgres server available on :" + testPort)
	}

	ctx := context.Background()
	collector := NewCollector(dsn)
	require.NoError(t, collector.Start(ctx))
	defer func() { _ = collector.Cleanup() }()

	observation, err := collector.Collect(ctx)
	require.NoError(t, err)
	require.NotNil(t, observation.Metrics)
	// numbackends is at least 1 (this connection).
	assert.NotEqual(t, uint64(0), observation.Metrics.Postgres.Backends)
	_ = observation
}

func TestObservedAttribution(t *testing.T) {
	t.Parallel()

	static := analysis.BuildGraph(
		[]analysis.Statement{{QueryID: 111, Query: "SELECT 1"}},
		[]analysis.Table{{Name: "users"}},
	)
	graph := model.BuildRuntimeGraph(analysis.ModulePath, static)

	profiles := trace.NewSnapshot(
		[]trace.Statement{{QueryID: 111, Query: "SELECT 1", Calls: 5, TotalExec: 2.5}},
		[]trace.Table{{Name: "users", SeqScan: 4, LiveTuples: 200}},
	).Profiles(time.Second)
	graph.ApplyUpdate(graph.BuildUpdate(&model.RuntimeMetrics{}, profiles, nil))

	byName := make(map[string]model.CodeMetrics)
	for _, node := range graph.Nodes {
		byName[node.Static.Name] = node.Metrics
	}

	statement := byName["SELECT 1"]
	require.NotNil(t, statement, "statement node 111 should have attributed heat")
	assert.Equal(t, int64(5), statement[model.Metric{Source: trace.StatementsSource, Name: "calls", Unit: "count"}].Self)

	users := byName["users"]
	require.NotNil(t, users, "table node users should have attributed heat")
	assert.Equal(t, int64(4), users[model.Metric{Source: trace.TablesSource, Name: "operations", Unit: "count"}].Self)
}

func TestStopUnblocksWaitAndCleanupIsIdempotent(t *testing.T) {
	collector := NewCollector("postgres://unused")
	done := make(chan error, 1)
	go func() { done <- collector.Wait() }()

	require.NoError(t, collector.Stop())
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Wait did not unblock after Stop")
	}
	require.NoError(t, collector.Cleanup())
}
