// Package trace defines Senbon's representation of a Go runtime trace.
package trace

import "time"

// EventKind identifies a runtime trace event.
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

// State is a goroutine or processor execution state.
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

// StackID identifies a deduplicated stack in a trace.
type StackID uint64

// Trace contains decoded events and their shared stacks.
type Trace struct {
	Duration time.Duration
	Events   []Event
	Stacks   map[StackID]Stack
}

// Event contains the common and kind-specific data of a trace event.
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

// Resource identifies the goroutine, processor, or thread affected by an event.
type Resource struct {
	Kind ResourceKind
	ID   int64
}

// Stack is a sequence of call frames, starting with the leaf frame.
type Stack struct {
	Frames []Frame
}

// Frame identifies one function call in a stack.
type Frame struct {
	PC       uint64
	Function string
	File     string
	Line     uint64
}
