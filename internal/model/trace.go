package model

import (
	"sync"
	"time"
)

const maxPooledResources = 4096

type resourceState struct {
	state State
	since time.Duration
	stack StackID
	live  bool
}

type rangeKey struct {
	kind ResourceKind
	id   int64
	name string
}

type traceWorkspace struct {
	goroutines map[int64]*resourceState
	processors map[int64]*resourceState
	threads    map[int64]struct{}
	ranges     map[rangeKey]time.Duration
}

var (
	traceWorkspaces = sync.Pool{New: func() any {
		return &traceWorkspace{
			goroutines: make(map[int64]*resourceState),
			processors: make(map[int64]*resourceState),
			threads:    make(map[int64]struct{}),
			ranges:     make(map[rangeKey]time.Duration),
		}
	}}
	resourceStates = sync.Pool{New: func() any { return new(resourceState) }}
)

func (g *RuntimeGraph) buildTrace(trace *Trace) *TraceUpdate {
	update := acquireTraceUpdate()
	if trace == nil {
		return update
	}
	update.Summary.Duration = trace.Duration
	workspace := acquireTraceWorkspace()
	targets := acquireTargetWorkspace()
	defer releaseTraceWorkspace(workspace)
	defer releaseTargetWorkspace(targets)
	var live uint64

	for _, event := range trace.Events {
		if event.Goroutine >= 0 {
			if _, ok := workspace.goroutines[event.Goroutine]; !ok {
				workspace.goroutines[event.Goroutine] = resourceStates.Get().(*resourceState)
			}
		}
		if event.Processor >= 0 {
			if _, ok := workspace.processors[event.Processor]; !ok {
				workspace.processors[event.Processor] = resourceStates.Get().(*resourceState)
			}
		}
		if event.Thread >= 0 {
			workspace.threads[event.Thread] = struct{}{}
		}

		switch event.Kind {
		case EventMetric:
			update.Summary.Metrics[event.Name] = event.Value
		case EventStackSample:
			update.Summary.StackSamples++
			nodes, files := g.mapper.traceTargets(trace, event.Stack, targets)
			update.Code.add(Metric{Source: TraceSource, Name: traceSamples, Unit: unitCount}, 1, nodes, files)
		case EventRangeBegin, EventRangeActive:
			workspace.ranges[rangeKey{kind: event.Resource.Kind, id: event.Resource.ID, name: event.Name}] = event.At
		case EventRangeEnd:
			key := rangeKey{kind: event.Resource.Kind, id: event.Resource.ID, name: event.Name}
			if start, ok := workspace.ranges[key]; ok && event.At >= start {
				update.Summary.Ranges[event.Name] += event.At - start
				delete(workspace.ranges, key)
			}
		case EventStateTransition:
			switch event.Resource.Kind {
			case ResourceGoroutine:
				state := workspace.goroutines[event.Resource.ID]
				if state == nil {
					state = resourceStates.Get().(*resourceState)
					workspace.goroutines[event.Resource.ID] = state
				}
				g.closeGoroutineState(update, trace, state, event.At, targets)
				if event.To == StateNotExist {
					if state.live {
						state.live = false
						live--
					}
				} else if !state.live {
					state.live = true
					live++
					if live > update.Summary.PeakGoroutines {
						update.Summary.PeakGoroutines = live
					}
				}
				state.state = event.To
				state.since = event.At
				state.stack = event.ResourceStack
				if state.stack == 0 && event.Goroutine == event.Resource.ID {
					state.stack = event.Stack
				}
			case ResourceProcessor:
				state := workspace.processors[event.Resource.ID]
				if state == nil {
					state = resourceStates.Get().(*resourceState)
					workspace.processors[event.Resource.ID] = state
				}
				closeProcessorState(&update.Summary, state, event.At)
				state.state = event.To
				state.since = event.At
			case ResourceThread:
				if event.Resource.ID >= 0 {
					workspace.threads[event.Resource.ID] = struct{}{}
				}
			}
		}
	}

	for _, state := range workspace.goroutines {
		g.closeGoroutineState(update, trace, state, trace.Duration, targets)
	}
	for _, state := range workspace.processors {
		closeProcessorState(&update.Summary, state, trace.Duration)
	}
	for key, start := range workspace.ranges {
		if trace.Duration >= start {
			update.Summary.Ranges[key.name] += trace.Duration - start
		}
	}
	update.Summary.Goroutines = uint64(len(workspace.goroutines))
	update.Summary.LiveGoroutines = live
	update.Summary.Processors = uint64(len(workspace.processors))
	update.Summary.Threads = uint64(len(workspace.threads))
	return update
}

func (g *RuntimeGraph) closeGoroutineState(update *TraceUpdate, trace *Trace, state *resourceState, at time.Duration, targets *targetWorkspace) {
	if at < state.since || state.state == StateUnknown || state.state == StateNotExist {
		return
	}
	duration := at - state.since
	update.Summary.GoroutineStates[state.state] += duration
	if duration == 0 || state.state != StateRunnable && state.state != StateWaiting && state.state != StateSyscall {
		return
	}
	nodes, files := g.mapper.traceTargets(trace, state.stack, targets)
	update.Code.add(Metric{Source: TraceSource, Name: string(state.state), Unit: unitNanoseconds}, int64(duration), nodes, files)
}

func closeProcessorState(summary *TraceSummary, state *resourceState, at time.Duration) {
	if at < state.since || state.state == StateUnknown || state.state == StateNotExist {
		return
	}
	summary.ProcessorStates[state.state] += at - state.since
}

func acquireTraceWorkspace() *traceWorkspace {
	return traceWorkspaces.Get().(*traceWorkspace)
}

func releaseTraceWorkspace(workspace *traceWorkspace) {
	for _, state := range workspace.goroutines {
		*state = resourceState{}
		resourceStates.Put(state)
	}
	for _, state := range workspace.processors {
		*state = resourceState{}
		resourceStates.Put(state)
	}
	if len(workspace.goroutines) > maxPooledResources {
		workspace.goroutines = make(map[int64]*resourceState)
	} else {
		clear(workspace.goroutines)
	}
	if len(workspace.processors) > maxPooledResources {
		workspace.processors = make(map[int64]*resourceState)
	} else {
		clear(workspace.processors)
	}
	if len(workspace.threads) > maxPooledResources {
		workspace.threads = make(map[int64]struct{})
	} else {
		clear(workspace.threads)
	}
	if len(workspace.ranges) > maxPooledResources {
		workspace.ranges = make(map[rangeKey]time.Duration)
	} else {
		clear(workspace.ranges)
	}
	traceWorkspaces.Put(workspace)
}
