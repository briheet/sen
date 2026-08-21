package model

import (
	"slices"
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

type mappedTraceStack struct {
	nodes []NodeID
	files []FileID
}

type traceWorkspace struct {
	goroutines map[int64]*resourceState
	processors map[int64]*resourceState
	threads    map[int64]struct{}
	ranges     map[rangeKey]time.Duration
	targets    map[StackID]mappedTraceStack
}

var (
	traceWorkspaces = sync.Pool{New: func() any {
		return &traceWorkspace{
			goroutines: make(map[int64]*resourceState),
			processors: make(map[int64]*resourceState),
			threads:    make(map[int64]struct{}),
			ranges:     make(map[rangeKey]time.Duration),
			targets:    make(map[StackID]mappedTraceStack),
		}
	}}
	resourceStates = sync.Pool{New: func() any { return new(resourceState) }}
)

func (g *RuntimeGraph) buildTrace(trace *Trace) *TraceUpdate {
	update := acquireTraceUpdate()
	if trace == nil {
		return update
	}
	if trace.Aggregate != nil {
		g.buildTraceAggregate(update, trace)
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
			nodes, files := g.traceTargets(trace, event.Stack, workspace, targets)
			update.Code.add(Metric{Source: TraceSource, Name: traceSamples, Unit: unitCount}, 1, nodes, files)
			addNodeTraceEdges(update.NodeEdges, nodes)
			addFileTraceEdges(update.FileEdges, files)
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
				g.closeGoroutineState(update, trace, state, event.At, workspace, targets)
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
		g.closeGoroutineState(update, trace, state, trace.Duration, workspace, targets)
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

func (g *RuntimeGraph) buildTraceAggregate(update *TraceUpdate, trace *Trace) {
	assignTraceSummary(&update.Summary, trace.Aggregate.Summary)
	workspace := acquireTraceWorkspace()
	targets := acquireTargetWorkspace()
	defer releaseTraceWorkspace(workspace)
	defer releaseTargetWorkspace(targets)

	for stackID, aggregate := range trace.Aggregate.Stacks {
		nodes, files := g.traceTargets(trace, stackID, workspace, targets)
		if aggregate.Samples != 0 {
			update.Code.add(Metric{Source: TraceSource, Name: traceSamples, Unit: unitCount}, aggregate.Samples, nodes, files)
			addNodeTraceEdgesN(update.NodeEdges, nodes, aggregate.Samples)
			addFileTraceEdgesN(update.FileEdges, files, aggregate.Samples)
		}
		g.addTraceStackCost(update, StateRunnable, aggregate.Runnable, nodes, files)
		g.addTraceStackCost(update, StateWaiting, aggregate.Waiting, nodes, files)
		g.addTraceStackCost(update, StateSyscall, aggregate.Syscall, nodes, files)
	}
}

func (g *RuntimeGraph) addTraceStackCost(update *TraceUpdate, state State, cost TraceStackCost, nodes []NodeID, files []FileID) {
	if cost.Duration == 0 {
		return
	}
	update.Code.add(Metric{Source: TraceSource, Name: string(state), Unit: unitNanoseconds}, int64(cost.Duration), nodes, files)
	addNodeTraceEdgesN(update.NodeEdges, nodes, cost.Occurrences)
	addFileTraceEdgesN(update.FileEdges, files, cost.Occurrences)
}

func (g *RuntimeGraph) traceTargets(trace *Trace, stackID StackID, workspace *traceWorkspace, targets *targetWorkspace) ([]NodeID, []FileID) {
	if mapped, ok := workspace.targets[stackID]; ok {
		return mapped.nodes, mapped.files
	}
	nodes, files := g.mapper.traceTargets(trace, stackID, targets)
	mapped := mappedTraceStack{nodes: slices.Clone(nodes), files: slices.Clone(files)}
	workspace.targets[stackID] = mapped
	return mapped.nodes, mapped.files
}

func addNodeTraceEdges(edges map[NodeEdge]int64, stack []NodeID) {
	addNodeTraceEdgesN(edges, stack, 1)
}

func addNodeTraceEdgesN(edges map[NodeEdge]int64, stack []NodeID, count int64) {
	for index := len(stack) - 1; index > 0; index-- {
		edges[NodeEdge{From: stack[index], To: stack[index-1]}] += count
	}
}

func addFileTraceEdges(edges map[FileEdge]int64, stack []FileID) {
	addFileTraceEdgesN(edges, stack, 1)
}

func addFileTraceEdgesN(edges map[FileEdge]int64, stack []FileID, count int64) {
	for index := len(stack) - 1; index > 0; index-- {
		edges[FileEdge{From: stack[index], To: stack[index-1]}] += count
	}
}

func (g *RuntimeGraph) closeGoroutineState(update *TraceUpdate, trace *Trace, state *resourceState, at time.Duration, workspace *traceWorkspace, targets *targetWorkspace) {
	if at < state.since || state.state == "" || state.state == StateUnknown || state.state == StateNotExist {
		return
	}
	duration := at - state.since
	update.Summary.GoroutineStates[state.state] += duration
	if duration == 0 || state.state != StateRunnable && state.state != StateWaiting && state.state != StateSyscall {
		return
	}
	nodes, files := g.traceTargets(trace, state.stack, workspace, targets)
	update.Code.add(Metric{Source: TraceSource, Name: string(state.state), Unit: unitNanoseconds}, int64(duration), nodes, files)
	addNodeTraceEdges(update.NodeEdges, nodes)
	addFileTraceEdges(update.FileEdges, files)
}

func closeProcessorState(summary *TraceSummary, state *resourceState, at time.Duration) {
	if at < state.since || state.state == "" || state.state == StateUnknown || state.state == StateNotExist {
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
	if len(workspace.targets) > maxPooledResources {
		workspace.targets = make(map[StackID]mappedTraceStack)
	} else {
		clear(workspace.targets)
	}
	traceWorkspaces.Put(workspace)
}
