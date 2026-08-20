package metrics

import (
	"testing"
	"time"

	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	body := "# Memory\r\n" +
		"used_memory:1048576\r\n" +
		"used_memory_peak:3145728\r\n" +
		"used_memory_dataset:524288\r\n" +
		"used_memory_rss:2097152\r\n" +
		"maxmemory:8388608\r\n" +
		"mem_fragmentation_ratio:2.0\r\n" +
		"# Stats\r\n" +
		"total_commands_processed:300\r\n" +
		"total_connections_received:12\r\n" +
		"instantaneous_ops_per_sec:25\r\n" +
		"total_net_input_bytes:4096\r\n" +
		"total_net_output_bytes:8192\r\n" +
		"keyspace_hits:90\r\n" +
		"keyspace_misses:10\r\n" +
		"expired_keys:7\r\n" +
		"evicted_keys:2\r\n" +
		"# CPU\r\n" +
		"used_cpu_user:0.5\r\n" +
		"used_cpu_sys:0.25\r\n" +
		"# Clients\r\n" +
		"connected_clients:4\r\n" +
		"blocked_clients:1\r\n" +
		"# Server\r\n" +
		"redis_version:8.0.0\r\n" +
		"redis_mode:standalone\r\n" +
		"role:master\r\n" +
		"uptime_in_seconds:120\r\n" +
		"# Keyspace\r\n" +
		"db0:keys=12,expires=2,avg_ttl=1000\r\n" +
		"db2:keys=3,expires=0,avg_ttl=0\r\n"

	m := Decode(body)

	assert.Equal(t, uint64(1048576), m.Redis.UsedMemory)
	assert.Equal(t, uint64(3145728), m.Redis.PeakMemory)
	assert.Equal(t, uint64(524288), m.Redis.UsedMemoryDataset)
	assert.Equal(t, uint64(2097152), m.Redis.RSS)
	assert.Equal(t, uint64(8388608), m.Redis.MaxMemory)
	assert.Equal(t, 2.0, m.Redis.MemoryFragmentationRatio)
	assert.Equal(t, uint64(300), m.Redis.TotalCommandsProcessed)
	assert.Equal(t, uint64(12), m.Redis.TotalConnectionsReceived)
	assert.Equal(t, uint64(4), m.Redis.ConnectedClients)
	assert.Equal(t, uint64(1), m.Redis.BlockedClients)
	assert.Equal(t, uint64(15), m.Redis.Keys)
	assert.Equal(t, uint64(25), m.Redis.InstantaneousOps)
	assert.Equal(t, uint64(4096), m.Redis.NetworkInputBytes)
	assert.Equal(t, uint64(8192), m.Redis.NetworkOutputBytes)
	assert.Equal(t, uint64(90), m.Redis.KeyspaceHits)
	assert.Equal(t, uint64(10), m.Redis.KeyspaceMisses)
	assert.Equal(t, uint64(7), m.Redis.ExpiredKeys)
	assert.Equal(t, uint64(2), m.Redis.EvictedKeys)
	assert.Equal(t, "8.0.0", m.Redis.Version)
	assert.Equal(t, "standalone", m.Redis.Mode)
	assert.Equal(t, "master", m.Redis.Role)
	assert.Equal(t, 120*time.Second, m.Redis.Uptime)
	assert.Equal(t, 0.5, m.Redis.UserCPU)
	assert.Equal(t, 0.25, m.Redis.SystemCPU)
	assert.Equal(t, uint64(2097152), m.Process.RSS)
	assert.True(t, m.Process.Has(model.ProcessCPU))
	assert.True(t, m.Process.Has(model.ProcessMemory))
	assert.True(t, m.Process.Has(model.ProcessUptime))
}

func TestDecodeMissingFields(t *testing.T) {
	t.Parallel()

	m := Decode("")
	assert.Equal(t, uint64(0), m.Redis.UsedMemory)
	assert.Equal(t, 0.0, m.Redis.UserCPU)
	assert.False(t, m.Process.Has(model.ProcessCPU))
	assert.False(t, m.Process.Has(model.ProcessMemory))
	assert.False(t, m.Process.Has(model.ProcessUptime))
}
