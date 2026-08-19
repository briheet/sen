// Package metrics defines sen's representation of Go runtime metrics.
package metrics

import "github.com/briheet/sen/internal/model"

const (
	// CPU and GC metrics.
	MetricsUserCPU  = "/cpu/classes/user:cpu-seconds"
	MetricsGCCPU    = "/cpu/classes/gc/total:cpu-seconds"
	MetricsGCAssist = "/cpu/classes/gc/mark/assist:cpu-seconds"
	MetricsGCCycles = "/gc/cycles/total:gc-cycles"

	// Heap metrics.
	MetricsHeapAlloc       = "/gc/heap/allocs:bytes"
	MetricsAllocCount      = "/gc/heap/allocs:objects"
	MetricsLiveHeap        = "/gc/heap/live:bytes"
	MetricsCurrHeapObjects = "/gc/heap/objects:objects"
	MetricsHeapGoal        = "/gc/heap/goal:bytes"
	MetricsMemoryLimit     = "/gc/gomemlimit:bytes"
	MetricsGOGC            = "/gc/gogc:percent"

	// Memory metrics.
	MetricsTotalRuntimeMem = "/memory/classes/total:bytes"
	MetricsStackMemory     = "/memory/classes/heap/stacks:bytes"
	MetricsHeapReleased    = "/memory/classes/heap/released:bytes"
	MetricsHeapFree        = "/memory/classes/heap/free:bytes"
	MetricsHeapUnused      = "/memory/classes/heap/unused:bytes"

	// Scheduler metrics.
	MetricsTotalLiveGoroutines = "/sched/goroutines:goroutines"
	MetricsGOMAXPROCS          = "/sched/gomaxprocs:threads"

	// Histogram metrics.
	MetricsSchedulerLatency = "/sched/latencies:seconds"
	MetricsGCPauses         = "/sched/pauses/total/gc:seconds"

	// Contention metrics.
	MetricsLockContention = "/sync/mutex/wait/total:seconds"
)

// RuntimeMetrics is the normalized process metric snapshot.
type RuntimeMetrics = model.RuntimeMetrics
