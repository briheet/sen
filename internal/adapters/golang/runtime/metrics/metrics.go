// Package metrics defines sen's representation of Go runtime metrics.
package metrics

import "github.com/briheet/sen/internal/model"

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

// RuntimeMetrics is the normalized process metric snapshot.
type RuntimeMetrics = model.RuntimeMetrics
