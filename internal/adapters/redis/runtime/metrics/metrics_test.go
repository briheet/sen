package metrics

import (
	"testing"

	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	body := "# Memory\r\n" +
		"used_memory:1048576\r\n" +
		"used_memory_dataset:524288\r\n" +
		"used_memory_rss:2097152\r\n" +
		"# Stats\r\n" +
		"total_commands_processed:300\r\n" +
		"total_connections_received:12\r\n" +
		"# CPU\r\n" +
		"used_cpu_user:0.5\r\n" +
		"used_cpu_sys:0.25\r\n" +
		"# Clients\r\n" +
		"connected_clients:4\r\n"

	m := Decode(body)

	assert.Equal(t, uint64(1048576), m.Redis.UsedMemory)
	assert.Equal(t, uint64(524288), m.Redis.UsedMemoryDataset)
	assert.Equal(t, uint64(2097152), m.Redis.RSS)
	assert.Equal(t, uint64(300), m.Redis.TotalCommandsProcessed)
	assert.Equal(t, uint64(12), m.Redis.TotalConnectionsReceived)
	assert.Equal(t, uint64(4), m.Redis.ConnectedClients)
	assert.Equal(t, 0.5, m.Redis.UserCPU)
	assert.Equal(t, 0.25, m.Redis.SystemCPU)
	assert.Equal(t, uint64(2097152), m.Process.RSS)
	assert.True(t, m.Process.Has(model.ProcessCPU))
	assert.True(t, m.Process.Has(model.ProcessMemory))
}

func TestDecodeMissingFields(t *testing.T) {
	t.Parallel()

	m := Decode("")
	assert.Equal(t, uint64(0), m.Redis.UsedMemory)
	assert.Equal(t, 0.0, m.Redis.UserCPU)
}
