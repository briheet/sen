// Package metrics defines Senbon's representation of Go runtime metrics.
package metrics

import runtimemetrics "runtime/metrics"

const (
	// CPU and GC metrics.
	MetricsUserCPU  = "/cpu/classes/user:cpu-seconds"
	MetricsGCCPU    = "/cpu/classes/gc/total:cpu-seconds"
	MetricsGCCycles = "/gc/cycles/total:gc-cycles"

	// Heap metrics.
	MetricsHeapAlloc       = "/gc/heap/allocs:bytes"
	MetricsAllocCount      = "/gc/heap/allocs:objects"
	MetricsLiveHeap        = "/gc/heap/live:bytes"
	MetricsCurrHeapObjects = "/gc/heap/objects:objects"

	// Memory metrics.
	MetricsTotalRuntimeMem = "/memory/classes/total:bytes"
	MetricsStackMemory     = "/memory/classes/heap/stacks:bytes"

	// Scheduler metrics.
	MetricsTotalLiveGoroutines = "/sched/goroutines:goroutines"
	MetricsRunningGoroutines   = "/sched/goroutines/running:goroutines"
	MetricsRunnableGoroutines  = "/sched/goroutines/runnable:goroutines"
	MetricsWaitingGoroutines   = "/sched/goroutines/waiting:goroutines"
	MetricsThreads             = "/sched/threads/total:threads"

	// Histogram metrics.
	MetricsSchedulerLatency = "/sched/latencies:seconds"
	MetricsGCPauses         = "/sched/pauses/total/gc:seconds"

	// Contention metrics.
	MetricsLockContention = "/sync/mutex/wait/total:seconds"
)

// RuntimeMetrics contains metrics collected from the target Go runtime.
type RuntimeMetrics struct {
	UserCPU  float64
	GCCPU    float64
	GCCycles uint64

	HeapAlloc       uint64
	AllocCount      uint64
	LiveHeap        uint64
	CurrHeapObjects uint64

	TotalRuntimeMem uint64
	StackMemory     uint64

	TotalLiveGoroutines uint64
	RunningGoroutines   uint64
	RunnableGoroutines  uint64
	WaitingGoroutines   uint64
	Threads             uint64

	SchedulerLatency *runtimemetrics.Float64Histogram
	GCPauses         *runtimemetrics.Float64Histogram

	LockContention float64
}
