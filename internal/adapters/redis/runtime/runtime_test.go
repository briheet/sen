package runtime

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testAddr returns a reachable Redis address or "" if none is running.
func testAddr(t *testing.T) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", "127.0.0.1:6379", 200*time.Millisecond)
	if err != nil {
		return ""
	}
	_ = conn.Close()
	return "127.0.0.1:6379"
}

func TestCollectorLive(t *testing.T) {
	addr := testAddr(t)
	if addr == "" {
		t.Skip("no redis server available at 127.0.0.1:6379")
	}

	ctx := context.Background()
	collector := NewCollector(addr)
	require.NoError(t, collector.Start(ctx))
	defer func() { _ = collector.Cleanup() }()

	observation, err := collector.Collect(ctx)
	require.NoError(t, err)
	require.NotNil(t, observation.Metrics)
	assert.NotEqual(t, uint64(0), observation.Metrics.Redis.UsedMemory)
	_ = observation
}
