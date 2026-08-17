// This file deals with runtime/metrics representation as conformed to go's runtime api
package runtime

const (
	// CPU and GC based runtime metrics
	MetricsUserCPU  = "/cpu/classes/user:cpu-seconds"
	MetricsGCCPU    = "/cpu/classes/gc/total:cpu-seconds"
	MetricsGCCycles = "/gc/cycles/total:gc-cycles"

	// Heap based metrics
	MetricsHeapAlloc       = "/gc/heap/allocs:bytes"
	MetricsAllocCount      = "/gc/heap/allocs:objects"
	MetricsLiveHeap        = "/gc/heap/live:bytes"
	MetricsCurrHeapObjects = "/gc/heap/objects:objects"

	// Memory based metrics
	MetricsTotalRuntimeMem = "/memory/classes/total:bytes"
	MetricsStackMemory     = "/memory/classes/heap/stacks:bytes"

	// Scheduler based metrics
	MetricsTotalLiveGoroutines = "/sched/goroutines:goroutines"
	MetricsRunningGoroutines   = "/sched/goroutines/running:goroutines"
	MetricsRunnableGoroutines  = "/sched/goroutines/runnable:goroutines"
	MetricsWaitingGoroutines   = "/sched/goroutines/waiting:goroutines"
	MetricsThreads             = "/sched/threads/total:threads"

	// Histograms based metrics
	MetricsSchedulerLatency = "/sched/latencies:seconds"
	MetricsGCPauses         = "/sched/pauses/total/gc:seconds"

	// Lock contention based metrics
	MetricsLockContention = "/sync/mutex/wait/total:seconds"
)

// This struct deals with Runtime metrics for internal representation
// conforming to go's runtime exposed api;s
// Any changes to runtime/metrics should be written edited here.
// Metrics are designed by string key. Its split into 2 via ":"
// The current supported ones are:
//
// 1.  /cpu/classes/user:cpu-seconds - user Go CPU time
// 2.  /cpu/classes/gc/total:cpu-seconds - GC CPU overhead
// 3.  /gc/cycles/total:gc-cycles - total GC cycles
// 4.  /gc/heap/allocs:bytes - acc. heap allocation
// 5.  /gc/heap/allocs:objects - allocation count
// 6.  /gc/heap/live:bytes — live heap
// 7.  /gc/heap/objects:objects — current heap objects
// 8.  /memory/classes/total:bytes — total Go runtime mapped memory
// 9.  /memory/classes/heap/stacks:bytes — goroutine stack memory
// 10. /sched/goroutines:goroutines — live goroutines
// 11. /sched/goroutines/running:goroutines — currently running
// 12. /sched/goroutines/runnable:goroutines — waiting for CPU
// 13. /sched/goroutines/waiting:goroutines — blocked/waiting
// 14. /sched/latencies:seconds — scheduler latency distribution
// 15. /sched/threads/total:threads — runtime OS threads
// 16. /sched/pauses/total/gc:seconds — GC STW pauses
// 17. /sync/mutex/wait/total:seconds — global lock contention
type RuntimeMetrics struct {
	// CPU / GC
	UserCPU  float64
	GCCPU    float64
	GCCycles uint64

	// Heap
	HeapAlloc       uint64
	AllocCount      uint64
	LiveHeap        uint64
	CurrHeapObjects uint64

	// Memory
	TotalRuntimeMem uint64
	StackMemory     uint64

	// Scheduler
	TotalLiveGoroutines uint64
	RunningGoroutines   uint64
	RunnableGoroutines  uint64
	WaitingGoroutines   uint64
	Threads             uint64

	// Histograms
	SchedulerLatency *metrics.Float64Histogram
	GCPauses         *metrics.Float64Histogram

	// Contention
	LockContention float64
}
