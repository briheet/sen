package runtime

import (
	"context"
	"net"
	"testing"
	"time"

	redistrace "github.com/briheet/sen/internal/adapters/redis/runtime/trace"
	"github.com/briheet/sen/internal/model"
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
	defer func() { _ = collector.Cleanup() }()
	if err := collector.Start(ctx); err != nil {
		t.Skipf("redis at %s is not healthy: %v", addr, err)
	}

	observation, err := collector.Collect(ctx)
	require.NoError(t, err)
	require.NotNil(t, observation.Metrics)
	assert.NotEqual(t, uint64(0), observation.Metrics.Redis.UsedMemory)

	_ = collector.client.Get(ctx, "sen:collector-test:missing").Err()
	observation, err = collector.Collect(ctx)
	require.NoError(t, err)
	require.Contains(t, observation.Profiles, redistrace.ProfileName)
	require.True(t, profileContainsCommand(observation.Profiles[redistrace.ProfileName], "GET"))
}

func profileContainsCommand(profile *model.Profile, command string) bool {
	if profile == nil {
		return false
	}
	for _, location := range profile.Locations {
		for _, frame := range location.Frames {
			if frame.Function == command {
				return true
			}
		}
	}
	return false
}

func TestStopUnblocksWaitAndCleanupIsIdempotent(t *testing.T) {
	collector := NewCollector("127.0.0.1:6379")
	done := make(chan error, 1)
	go func() { done <- collector.Wait() }()

	require.NoError(t, collector.Stop())
	require.NoError(t, <-done)
	require.NoError(t, collector.Cleanup())
}
