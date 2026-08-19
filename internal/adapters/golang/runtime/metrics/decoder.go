package metrics

import (
	"encoding/gob"
	"io"
	runtimemetrics "runtime/metrics"

	"github.com/briheet/sen/internal/model"
)

type wireMetric struct {
	Name      string
	Uint64    uint64
	Float64   float64
	Histogram *runtimemetrics.Float64Histogram
}

// Read decodes metrics sent by the injected collector.
func Read(reader io.Reader) (*RuntimeMetrics, error) {
	var samples []wireMetric
	if err := gob.NewDecoder(reader).Decode(&samples); err != nil {
		return nil, err
	}

	result := &RuntimeMetrics{}
	for _, sample := range samples {
		switch sample.Name {
		case MetricsUserCPU:
			result.Go.UserCPU = sample.Float64
		case MetricsGCCPU:
			result.Go.GCCPU = sample.Float64
		case MetricsGCAssist:
			result.Go.GCAssist = sample.Float64
		case MetricsGCCycles:
			result.Go.GCCycles = sample.Uint64
		case MetricsHeapAlloc:
			result.Go.AllocatedBytes = sample.Uint64
		case MetricsAllocCount:
			result.Go.Allocations = sample.Uint64
		case MetricsLiveHeap:
			result.Go.LiveHeap = sample.Uint64
		case MetricsCurrHeapObjects:
			result.Go.HeapObjects = sample.Uint64
		case MetricsHeapGoal:
			result.Go.HeapGoal = sample.Uint64
		case MetricsMemoryLimit:
			result.Go.MemoryLimit = sample.Uint64
		case MetricsGOGC:
			result.Go.GOGC = sample.Uint64
		case MetricsTotalRuntimeMem:
			result.Go.RuntimeMemory = sample.Uint64
		case MetricsStackMemory:
			result.Go.StackMemory = sample.Uint64
		case MetricsHeapReleased:
			result.Go.HeapReleased = sample.Uint64
		case MetricsHeapFree:
			result.Go.HeapFree = sample.Uint64
		case MetricsHeapUnused:
			result.Go.HeapUnused = sample.Uint64
		case MetricsTotalLiveGoroutines:
			result.Go.Goroutines = sample.Uint64
		case MetricsGOMAXPROCS:
			result.Go.GOMAXPROCS = sample.Uint64
		case MetricsSchedulerLatency:
			if sample.Histogram == nil {
				continue
			}
			result.Go.SchedulerLatency = &model.Histogram{
				Counts:  append([]uint64(nil), sample.Histogram.Counts...),
				Buckets: append([]float64(nil), sample.Histogram.Buckets...),
			}
		case MetricsGCPauses:
			if sample.Histogram == nil {
				continue
			}
			result.Go.GCPauses = &model.Histogram{
				Counts:  append([]uint64(nil), sample.Histogram.Counts...),
				Buckets: append([]float64(nil), sample.Histogram.Buckets...),
			}
		case MetricsLockContention:
			result.Go.MutexWait = sample.Float64
		}
	}
	return result, nil
}
