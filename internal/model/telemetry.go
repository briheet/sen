package model

import "time"

// Histogram contains bucket boundaries and their counts.
type Histogram struct {
	Counts  []uint64
	Buckets []float64
}

// RuntimeMetrics contains process-wide runtime measurements.
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

	SchedulerLatency *Histogram
	GCPauses         *Histogram
	LockContention   float64
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
