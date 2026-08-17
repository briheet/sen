package model

import "time"

// Observation is one normalized runtime snapshot from an adapter.
type Observation struct {
	Metrics  *RuntimeMetrics // Required for every observation.
	Profiles map[string]*Profile
	Trace    *Trace
}

// Histogram contains bucket boundaries and their counts.
type Histogram struct {
	Counts  []uint64
	Buckets []float64
}

// ProcessMetric identifies an OS measurement available for a target process.
type ProcessMetric uint16

const (
	ProcessCPU ProcessMetric = 1 << iota
	ProcessMemory
	ProcessThreads
	ProcessOpenFiles
	ProcessIO
	ProcessIOOperations
	ProcessContextSwitches
	ProcessUptime
)

// ProcessMetrics contains language-independent operating-system measurements.
type ProcessMetrics struct {
	UserCPU   float64
	SystemCPU float64
	Uptime    time.Duration

	RSS           uint64
	PeakRSS       uint64
	VirtualMemory uint64
	ReadBytes     uint64
	WriteBytes    uint64
	ReadOps       uint64
	WriteOps      uint64
	VoluntaryCS   uint64
	InvoluntaryCS uint64
	Threads       uint64
	OpenFiles     uint64

	Available ProcessMetric
}

// Has reports whether a process measurement was collected on this platform.
func (m ProcessMetrics) Has(metric ProcessMetric) bool { return m.Available&metric != 0 }

// GoMetrics contains measurements exposed by the Go runtime.
type GoMetrics struct {
	UserCPU  float64
	GCCPU    float64
	GCAssist float64
	GCCycles uint64

	AllocatedBytes uint64
	Allocations    uint64
	LiveHeap       uint64
	HeapObjects    uint64
	HeapGoal       uint64
	MemoryLimit    uint64
	RuntimeMemory  uint64
	StackMemory    uint64
	HeapReleased   uint64
	HeapFree       uint64
	HeapUnused     uint64
	GOGC           uint64
	Goroutines     uint64
	GOMAXPROCS     uint64

	SchedulerLatency *Histogram
	GCPauses         *Histogram
	MutexWait        float64
}

// NodeMetrics contains measurements exposed by the Node.js runtime.
type NodeMetrics struct {
	HeapUsed     uint64
	HeapTotal    uint64
	External     uint64
	ArrayBuffers uint64

	EventLoopUtilization float64
	EventLoopDelayMean   time.Duration
	EventLoopDelayMax    time.Duration
	EventLoopDelayP95    time.Duration
	EventLoopDelayP99    time.Duration
	ActiveResources      uint64
}

// RedisMetrics contains measurements exposed by a Redis server.
type RedisMetrics struct {
	UsedMemory               uint64
	UsedMemoryDataset        uint64
	RSS                      uint64
	UserCPU                  float64
	SystemCPU                float64
	ConnectedClients         uint64
	TotalCommandsProcessed   uint64
	TotalConnectionsReceived uint64
}

// RuntimeMetrics contains one normalized process and runtime snapshot.
type RuntimeMetrics struct {
	Process ProcessMetrics
	Go      GoMetrics
	Node    NodeMetrics
	Redis   RedisMetrics
}

// EventKind identifies a runtime event.
type EventKind string

const (
	EventSync            EventKind = "sync"
	EventMetric          EventKind = "metric"
	EventLabel           EventKind = "label"
	EventStackSample     EventKind = "stack-sample"
	EventRangeBegin      EventKind = "range-begin"
	EventRangeActive     EventKind = "range-active"
	EventRangeEnd        EventKind = "range-end"
	EventTaskBegin       EventKind = "task-begin"
	EventTaskEnd         EventKind = "task-end"
	EventRegionBegin     EventKind = "region-begin"
	EventRegionEnd       EventKind = "region-end"
	EventLog             EventKind = "log"
	EventStateTransition EventKind = "state-transition"
)

// State identifies an execution state.
type State string

const (
	StateUnknown  State = "unknown"
	StateNotExist State = "not-exist"
	StateRunnable State = "runnable"
	StateRunning  State = "running"
	StateWaiting  State = "waiting"
	StateSyscall  State = "syscall"
	StateIdle     State = "idle"
)

// ResourceKind identifies a runtime resource.
type ResourceKind string

const (
	ResourceNone      ResourceKind = "none"
	ResourceGoroutine ResourceKind = "goroutine"
	ResourceProcessor ResourceKind = "processor"
	ResourceThread    ResourceKind = "thread"
)

// StackID identifies a deduplicated trace stack.
type StackID uint64

// Trace contains decoded events and their shared stacks.
type Trace struct {
	Duration time.Duration
	Events   []Event
	Stacks   map[StackID]TraceStack
}

// Event contains common and event-specific runtime data.
type Event struct {
	At        time.Duration
	Kind      EventKind
	Goroutine int64
	Processor int64
	Thread    int64
	Stack     StackID

	Resource      Resource
	ResourceStack StackID
	From          State
	To            State
	Reason        string

	Task     uint64
	Parent   uint64
	Name     string
	Category string
	Message  string
	Value    uint64
}

// Resource identifies a runtime resource affected by an event.
type Resource struct {
	Kind ResourceKind
	ID   int64
}

// TraceStack is a leaf-first sequence of runtime frames.
type TraceStack struct {
	Frames []TraceFrame
}

// TraceFrame identifies one function call in a runtime stack.
type TraceFrame struct {
	PC       uint64
	Function string
	File     string
	Line     uint64
}
