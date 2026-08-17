// Package metrics parses Redis INFO output into normalized runtime metrics.
package metrics

import (
	"strconv"
	"strings"

	"github.com/briheet/sen/internal/model"
)

// parse maps an INFO body (lines of "key:value" grouped by section) into a
// section-key lookup consumed by the field extractors.
func parse(body string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		result[key] = value
	}
	return result
}

func uintValue(fields map[string]string, key string) uint64 {
	value, err := strconv.ParseUint(fields[key], 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func floatValue(fields map[string]string, key string) float64 {
	value, err := strconv.ParseFloat(fields[key], 64)
	if err != nil {
		return 0
	}
	return value
}

// Decode translates one INFO snapshot into normalized runtime metrics.
// Callers should pass the body of `INFO memory stats cpu clients`.
func Decode(body string) *model.RuntimeMetrics {
	fields := parse(body)
	userCPU := floatValue(fields, "used_cpu_user")
	systemCPU := floatValue(fields, "used_cpu_sys")
	rss := uintValue(fields, "used_memory_rss")
	return &model.RuntimeMetrics{
		Process: model.ProcessMetrics{
			UserCPU:   userCPU,
			SystemCPU: systemCPU,
			RSS:       rss,
			PeakRSS:   rss,
			Available: model.ProcessCPU | model.ProcessMemory,
		},
		Redis: model.RedisMetrics{
			UsedMemory:               uintValue(fields, "used_memory"),
			UsedMemoryDataset:        uintValue(fields, "used_memory_dataset"),
			RSS:                      rss,
			UserCPU:                  userCPU,
			SystemCPU:                systemCPU,
			ConnectedClients:         uintValue(fields, "connected_clients"),
			TotalCommandsProcessed:   uintValue(fields, "total_commands_processed"),
			TotalConnectionsReceived: uintValue(fields, "total_connections_received"),
		},
	}
}
