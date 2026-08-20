// Package metrics translates Redis INFO output into sen's runtime model.
package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/briheet/sen/internal/model"
)

// infoFields is the flat key/value representation used by Redis INFO. Section
// headings are deliberately ignored because INFO keys are globally unique.
type infoFields map[string]string

func parse(body string) infoFields {
	result := make(infoFields)
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

func (fields infoFields) uint(key string) uint64 {
	value, err := strconv.ParseUint(fields[key], 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func (fields infoFields) float(key string) float64 {
	value, err := strconv.ParseFloat(fields[key], 64)
	if err != nil {
		return 0
	}
	return value
}

func (fields infoFields) has(keys ...string) bool {
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return false
		}
	}
	return true
}

func (fields infoFields) keys() uint64 {
	var total uint64
	for name, value := range fields {
		if !strings.HasPrefix(name, "db") {
			continue
		}
		for field := range strings.SplitSeq(value, ",") {
			key, raw, ok := strings.Cut(field, "=")
			if ok && key == "keys" {
				count, _ := strconv.ParseUint(raw, 10, 64)
				total += count
				break
			}
		}
	}
	return total
}

// Decode translates a complete INFO response into normalized runtime metrics.
func Decode(body string) *model.RuntimeMetrics {
	fields := parse(body)
	userCPU := fields.float("used_cpu_user")
	systemCPU := fields.float("used_cpu_sys")
	rss := fields.uint("used_memory_rss")
	uptime := time.Duration(fields.uint("uptime_in_seconds")) * time.Second
	var available model.ProcessMetric
	if fields.has("used_cpu_user", "used_cpu_sys") {
		available |= model.ProcessCPU
	}
	if fields.has("used_memory_rss") {
		available |= model.ProcessMemory
	}
	if fields.has("uptime_in_seconds") {
		available |= model.ProcessUptime
	}
	return &model.RuntimeMetrics{
		Process: model.ProcessMetrics{
			UserCPU:   userCPU,
			SystemCPU: systemCPU,
			RSS:       rss,
			Uptime:    uptime,
			Available: available,
		},
		Redis: model.RedisMetrics{
			Version:                  fields["redis_version"],
			Mode:                     fields["redis_mode"],
			Role:                     fields["role"],
			Uptime:                   uptime,
			UsedMemory:               fields.uint("used_memory"),
			PeakMemory:               fields.uint("used_memory_peak"),
			UsedMemoryDataset:        fields.uint("used_memory_dataset"),
			RSS:                      rss,
			MaxMemory:                fields.uint("maxmemory"),
			MemoryFragmentationRatio: fields.float("mem_fragmentation_ratio"),
			UserCPU:                  userCPU,
			SystemCPU:                systemCPU,
			ConnectedClients:         fields.uint("connected_clients"),
			BlockedClients:           fields.uint("blocked_clients"),
			Keys:                     fields.keys(),
			InstantaneousOps:         fields.uint("instantaneous_ops_per_sec"),
			TotalCommandsProcessed:   fields.uint("total_commands_processed"),
			TotalConnectionsReceived: fields.uint("total_connections_received"),
			NetworkInputBytes:        fields.uint("total_net_input_bytes"),
			NetworkOutputBytes:       fields.uint("total_net_output_bytes"),
			KeyspaceHits:             fields.uint("keyspace_hits"),
			KeyspaceMisses:           fields.uint("keyspace_misses"),
			ExpiredKeys:              fields.uint("expired_keys"),
			EvictedKeys:              fields.uint("evicted_keys"),
		},
	}
}
