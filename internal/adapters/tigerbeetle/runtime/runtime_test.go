package runtime

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testCluster = "0123456789abcdef0123456789abcdef"

func TestCollectorReadsOneBurst(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	collector := NewCollector("127.0.0.1:0", &output)
	require.NoError(t, collector.Start(context.Background()))
	t.Cleanup(func() { _ = collector.Cleanup() })

	client, err := net.Dial("udp", collector.conn.LocalAddr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	packet := "tb.replica_status:0|g|#cluster:" + testCluster + ",replica:0\n" +
		"tb.replica_view:7|g|#cluster:" + testCluster + ",replica:0\n" +
		"tb.grid_cache_hits:90|g|#cluster:" + testCluster + ",replica:0\n" +
		"tb.grid_cache_misses:10|g|#cluster:" + testCluster + ",replica:0\n" +
		"tb.replica_request_us.count:4|c|#cluster:" + testCluster + ",replica:0,operation:create_accounts\n" +
		"tb.replica_request_us.sum:100|c|#cluster:" + testCluster + ",replica:0,operation:create_accounts\n" +
		"tb.replica_request_us.min:10|g|#cluster:" + testCluster + ",replica:0,operation:create_accounts\n" +
		"tb.replica_request_us.max:50|g|#cluster:" + testCluster + ",replica:0,operation:create_accounts"
	_, err = client.Write([]byte(packet))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	observation, err := collector.Collect(ctx)
	require.NoError(t, err)
	require.Equal(t, testCluster, observation.Metrics.TigerBeetle.Cluster)
	require.Equal(t, uint64(7), observation.Metrics.TigerBeetle.Replicas[0].View)
	operation := observation.Metrics.TigerBeetle.Operations["create_accounts"]
	require.Equal(t, uint64(4), operation.Requests)
	require.Equal(t, 25*time.Microsecond, operation.LatencyAvg)
	require.Equal(t, 50*time.Microsecond, operation.LatencyMax)
	require.Len(t, observation.Profiles[profileName].Samples, 1)
}

func TestCollectorCollectHonorsCancellation(t *testing.T) {
	t.Parallel()
	collector := NewCollector("127.0.0.1:0", nil)
	require.NoError(t, collector.Start(context.Background()))
	t.Cleanup(func() { _ = collector.Cleanup() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := collector.Collect(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestCollectorStopUnblocksWait(t *testing.T) {
	t.Parallel()
	collector := NewCollector("127.0.0.1:0", nil)
	require.NoError(t, collector.Start(context.Background()))
	require.NoError(t, collector.Stop())
	require.NoError(t, collector.Cleanup())
	require.NoError(t, collector.Wait())
}
