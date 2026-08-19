// Package trace decodes runtime traces produced by a Go target.
package trace

import "github.com/briheet/sen/internal/model"

type (
	EventKind    = model.EventKind
	State        = model.State
	ResourceKind = model.ResourceKind
	StackID      = model.StackID
	Trace        = model.Trace
	Event        = model.Event
	Resource     = model.Resource
	Stack        = model.TraceStack
	Frame        = model.TraceFrame
)

const (
	EventSync            = model.EventSync
	EventMetric          = model.EventMetric
	EventLabel           = model.EventLabel
	EventStackSample     = model.EventStackSample
	EventRangeBegin      = model.EventRangeBegin
	EventRangeActive     = model.EventRangeActive
	EventRangeEnd        = model.EventRangeEnd
	EventTaskBegin       = model.EventTaskBegin
	EventTaskEnd         = model.EventTaskEnd
	EventRegionBegin     = model.EventRegionBegin
	EventRegionEnd       = model.EventRegionEnd
	EventLog             = model.EventLog
	EventStateTransition = model.EventStateTransition
	StateUnknown         = model.StateUnknown
	StateNotExist        = model.StateNotExist
	StateRunnable        = model.StateRunnable
	StateRunning         = model.StateRunning
	StateWaiting         = model.StateWaiting
	StateSyscall         = model.StateSyscall
	StateIdle            = model.StateIdle
	ResourceNone         = model.ResourceNone
	ResourceGoroutine    = model.ResourceGoroutine
	ResourceProcessor    = model.ResourceProcessor
	ResourceThread       = model.ResourceThread
)
