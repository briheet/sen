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
			result.UserCPU = sample.Float64
		case MetricsGCCPU:
			result.GCCPU = sample.Float64
		case MetricsGCCycles:
			result.GCCycles = sample.Uint64
		case MetricsHeapAlloc:
			result.HeapAlloc = sample.Uint64
		case MetricsAllocCount:
			result.AllocCount = sample.Uint64
		case MetricsLiveHeap:
			result.LiveHeap = sample.Uint64
		case MetricsCurrHeapObjects:
			result.CurrHeapObjects = sample.Uint64
		case MetricsTotalRuntimeMem:
			result.TotalRuntimeMem = sample.Uint64
		case MetricsStackMemory:
			result.StackMemory = sample.Uint64
		case MetricsTotalLiveGoroutines:
			result.TotalLiveGoroutines = sample.Uint64
		case MetricsRunningGoroutines:
			result.RunningGoroutines = sample.Uint64
		case MetricsRunnableGoroutines:
			result.RunnableGoroutines = sample.Uint64
		case MetricsWaitingGoroutines:
			result.WaitingGoroutines = sample.Uint64
		case MetricsThreads:
			result.Threads = sample.Uint64
		case MetricsSchedulerLatency:
			result.SchedulerLatency = &model.Histogram{
				Counts:  append([]uint64(nil), sample.Histogram.Counts...),
				Buckets: append([]float64(nil), sample.Histogram.Buckets...),
			}
		case MetricsGCPauses:
			result.GCPauses = &model.Histogram{
				Counts:  append([]uint64(nil), sample.Histogram.Counts...),
				Buckets: append([]float64(nil), sample.Histogram.Buckets...),
			}
		case MetricsLockContention:
			result.LockContention = sample.Float64
		}
	}
	return result, nil
}
