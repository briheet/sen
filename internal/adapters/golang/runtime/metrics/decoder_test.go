package metrics

import (
	"bytes"
	"encoding/gob"
	runtimemetrics "runtime/metrics"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadNormalizesGoRuntimeMetrics(t *testing.T) {
	histogram := &runtimemetrics.Float64Histogram{
		Counts:  []uint64{2, 1},
		Buckets: []float64{0, 0.001, 0.01},
	}
	samples := []wireMetric{
		{Name: MetricsUserCPU, Float64: 1.5},
		{Name: MetricsGCAssist, Float64: 0.25},
		{Name: MetricsGCCycles, Uint64: 4},
		{Name: MetricsLiveHeap, Uint64: 1024},
		{Name: MetricsHeapGoal, Uint64: 2048},
		{Name: MetricsMemoryLimit, Uint64: 4096},
		{Name: MetricsGOGC, Uint64: 100},
		{Name: MetricsHeapReleased, Uint64: 512},
		{Name: MetricsHeapFree, Uint64: 256},
		{Name: MetricsHeapUnused, Uint64: 128},
		{Name: MetricsGOMAXPROCS, Uint64: 8},
		{Name: MetricsSchedulerLatency, Histogram: histogram},
	}
	var encoded bytes.Buffer
	require.NoError(t, gob.NewEncoder(&encoded).Encode(samples))

	result, err := Read(&encoded)
	require.NoError(t, err)
	require.Equal(t, 1.5, result.Go.UserCPU)
	require.Equal(t, 0.25, result.Go.GCAssist)
	require.Equal(t, uint64(4), result.Go.GCCycles)
	require.Equal(t, uint64(1024), result.Go.LiveHeap)
	require.Equal(t, uint64(2048), result.Go.HeapGoal)
	require.Equal(t, uint64(4096), result.Go.MemoryLimit)
	require.Equal(t, uint64(100), result.Go.GOGC)
	require.Equal(t, uint64(512), result.Go.HeapReleased)
	require.Equal(t, uint64(256), result.Go.HeapFree)
	require.Equal(t, uint64(128), result.Go.HeapUnused)
	require.Equal(t, uint64(8), result.Go.GOMAXPROCS)
	require.Equal(t, []uint64{2, 1}, result.Go.SchedulerLatency.Counts)

	// Decoded histograms own their slices independently of the wire value.
	histogram.Counts[0] = 99
	require.Equal(t, uint64(2), result.Go.SchedulerLatency.Counts[0])
}

func TestReadIgnoresMissingHistogramPayload(t *testing.T) {
	var encoded bytes.Buffer
	require.NoError(t, gob.NewEncoder(&encoded).Encode([]wireMetric{{Name: MetricsGCPauses}}))

	result, err := Read(&encoded)
	require.NoError(t, err)
	require.Nil(t, result.Go.GCPauses)
}
