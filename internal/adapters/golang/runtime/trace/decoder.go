package trace

import (
	"io"
	"time"

	exptrace "golang.org/x/exp/trace"
)

const (
	maxReusableResources = 4 * 1024
	maxReusableStacks    = 4 * 1024
	maxReusableFrames    = 1024
)

// Decoder reuses storage across sequential runtime trace windows.
type Decoder struct {
	buffers     [2]decodeBuffer
	active      int
	initialized bool
}

type decodeBuffer struct {
	trace       Trace
	stackIDs    map[exptrace.Stack]StackID
	stackFrames [][]Frame
	stackCount  int
	goroutines  map[int64]decodeResourceState
	processors  map[int64]decodeResourceState
	threads     map[int64]struct{}
	ranges      map[decodeRangeKey]time.Duration
	live        uint64
}

type decodeResourceState struct {
	state State
	since time.Duration
	stack StackID
	live  bool
}

type decodeRangeKey struct {
	kind ResourceKind
	id   int64
	name string
}

// Read decodes a Go runtime trace into sen's representation.
func Read(r io.Reader) (*Trace, error) {
	var decoder Decoder
	return decoder.Read(r)
}

// Read decodes into the inactive buffer and publishes it only on success.
func (d *Decoder) Read(r io.Reader) (*Trace, error) {
	next := 0
	if d.initialized {
		next = 1 - d.active
	}
	buffer := &d.buffers[next]
	buffer.reset()
	if err := buffer.read(r); err != nil {
		return nil, err
	}
	d.active = next
	d.initialized = true
	return &buffer.trace, nil
}

func (b *decodeBuffer) reset() {
	b.trace.Duration = 0
	b.trace.Events = nil
	b.live = 0
	resetDecodeMap(&b.goroutines)
	resetDecodeMap(&b.processors)
	resetDecodeMap(&b.threads)
	resetDecodeMap(&b.ranges)
	if b.trace.Aggregate == nil {
		b.trace.Aggregate = newAggregate()
	} else {
		resetAggregate(b.trace.Aggregate)
	}
	if len(b.trace.Stacks) > maxReusableStacks {
		b.trace.Stacks = make(map[StackID]Stack)
		b.stackIDs = make(map[exptrace.Stack]StackID)
		b.stackFrames = nil
	} else {
		clear(b.trace.Stacks)
		clear(b.stackIDs)
		for index := range b.stackCount {
			if cap(b.stackFrames[index]) > maxReusableFrames {
				b.stackFrames[index] = nil
			} else {
				clear(b.stackFrames[index])
				b.stackFrames[index] = b.stackFrames[index][:0]
			}
		}
	}
	b.stackCount = 0
	if b.trace.Stacks == nil {
		b.trace.Stacks = make(map[StackID]Stack)
	}
	if b.stackIDs == nil {
		b.stackIDs = make(map[exptrace.Stack]StackID)
	}
}

func (b *decodeBuffer) read(r io.Reader) error {
	reader, err := exptrace.NewReader(r)
	if err != nil {
		return err
	}

	var (
		start   exptrace.Time
		started bool
	)

	for {
		source, err := reader.ReadEvent()
		if err == io.EOF {
			b.finish()
			return nil
		}
		if err != nil {
			return err
		}
		if source.Kind() == exptrace.EventExperimental {
			continue
		}
		if !started {
			start = source.Time()
			started = true
		}
		at := source.Time().Sub(start)
		b.trace.Duration = at
		b.trace.Aggregate.Summary.Duration = at
		b.observeResources(source)

		switch source.Kind() {
		case exptrace.EventMetric:
			metric := source.Metric()
			b.trace.Aggregate.Summary.Metrics[metric.Name] = metric.Value.Uint64()
		case exptrace.EventStackSample:
			stackID := b.addStack(source.Stack())
			metrics := b.trace.Aggregate.Stacks[stackID]
			metrics.Samples++
			b.trace.Aggregate.Stacks[stackID] = metrics
			b.trace.Aggregate.Summary.StackSamples++
		case exptrace.EventRangeBegin, exptrace.EventRangeActive:
			rangeEvent := source.Range()
			resource := resource(rangeEvent.Scope)
			b.ranges[decodeRangeKey{kind: resource.Kind, id: resource.ID, name: rangeEvent.Name}] = at
		case exptrace.EventRangeEnd:
			rangeEvent := source.Range()
			resource := resource(rangeEvent.Scope)
			key := decodeRangeKey{kind: resource.Kind, id: resource.ID, name: rangeEvent.Name}
			if since, ok := b.ranges[key]; ok && at >= since {
				b.trace.Aggregate.Summary.Ranges[rangeEvent.Name] += at - since
				delete(b.ranges, key)
			}
		case exptrace.EventStateTransition:
			b.transition(source, at)
		}
	}
}

func newAggregate() *Aggregate {
	return &Aggregate{
		Summary: Summary{
			GoroutineStates: make(map[State]time.Duration),
			ProcessorStates: make(map[State]time.Duration),
			Ranges:          make(map[string]time.Duration),
			Metrics:         make(map[string]uint64),
		},
		Stacks: make(map[StackID]StackMetrics),
	}
}

func resetAggregate(aggregate *Aggregate) {
	summary := &aggregate.Summary
	summary.Duration = 0
	summary.Goroutines = 0
	summary.LiveGoroutines = 0
	summary.PeakGoroutines = 0
	summary.Processors = 0
	summary.Threads = 0
	summary.StackSamples = 0
	resetDecodeMap(&summary.GoroutineStates)
	resetDecodeMap(&summary.ProcessorStates)
	resetDecodeMap(&summary.Ranges)
	resetDecodeMap(&summary.Metrics)
	resetDecodeMap(&aggregate.Stacks)
}

func resetDecodeMap[K comparable, V any](target *map[K]V) {
	if *target == nil || len(*target) > maxReusableResources {
		*target = make(map[K]V)
		return
	}
	clear(*target)
}

func (b *decodeBuffer) observeResources(source exptrace.Event) {
	if id := int64(source.Goroutine()); id >= 0 {
		if _, ok := b.goroutines[id]; !ok {
			b.goroutines[id] = decodeResourceState{state: StateUnknown}
		}
	}
	if id := int64(source.Proc()); id >= 0 {
		if _, ok := b.processors[id]; !ok {
			b.processors[id] = decodeResourceState{state: StateUnknown}
		}
	}
	if id := int64(source.Thread()); id >= 0 {
		b.threads[id] = struct{}{}
	}
}

func (b *decodeBuffer) transition(source exptrace.Event, at time.Duration) {
	transition := source.StateTransition()
	resource := resource(transition.Resource)
	_, to := states(transition)

	switch resource.Kind {
	case ResourceGoroutine:
		state, ok := b.goroutines[resource.ID]
		if !ok {
			state.state = StateUnknown
		}
		b.closeGoroutineState(state, at)
		if to == StateNotExist {
			if state.live {
				state.live = false
				b.live--
			}
		} else if !state.live {
			state.live = true
			b.live++
			if b.live > b.trace.Aggregate.Summary.PeakGoroutines {
				b.trace.Aggregate.Summary.PeakGoroutines = b.live
			}
		}
		state.state = to
		state.since = at
		state.stack = b.addStack(transition.Stack)
		if state.stack == 0 && int64(source.Goroutine()) == resource.ID {
			state.stack = b.addStack(source.Stack())
		}
		b.goroutines[resource.ID] = state
	case ResourceProcessor:
		state, ok := b.processors[resource.ID]
		if !ok {
			state.state = StateUnknown
		}
		b.closeProcessorState(state, at)
		state.state = to
		state.since = at
		b.processors[resource.ID] = state
	case ResourceThread:
		if resource.ID >= 0 {
			b.threads[resource.ID] = struct{}{}
		}
	}
}

func (b *decodeBuffer) closeGoroutineState(state decodeResourceState, at time.Duration) {
	if at < state.since || state.state == StateUnknown || state.state == StateNotExist {
		return
	}
	duration := at - state.since
	b.trace.Aggregate.Summary.GoroutineStates[state.state] += duration
	if duration == 0 {
		return
	}
	metrics := b.trace.Aggregate.Stacks[state.stack]
	var cost *StackCost
	switch state.state {
	case StateRunnable:
		cost = &metrics.Runnable
	case StateWaiting:
		cost = &metrics.Waiting
	case StateSyscall:
		cost = &metrics.Syscall
	default:
		return
	}
	cost.Duration += duration
	cost.Occurrences++
	b.trace.Aggregate.Stacks[state.stack] = metrics
}

func (b *decodeBuffer) closeProcessorState(state decodeResourceState, at time.Duration) {
	if at < state.since || state.state == StateUnknown || state.state == StateNotExist {
		return
	}
	b.trace.Aggregate.Summary.ProcessorStates[state.state] += at - state.since
}

func (b *decodeBuffer) finish() {
	for _, state := range b.goroutines {
		b.closeGoroutineState(state, b.trace.Duration)
	}
	for _, state := range b.processors {
		b.closeProcessorState(state, b.trace.Duration)
	}
	for key, since := range b.ranges {
		if b.trace.Duration >= since {
			b.trace.Aggregate.Summary.Ranges[key.name] += b.trace.Duration - since
		}
	}
	summary := &b.trace.Aggregate.Summary
	summary.Goroutines = uint64(len(b.goroutines))
	summary.LiveGoroutines = b.live
	summary.Processors = uint64(len(b.processors))
	summary.Threads = uint64(len(b.threads))
}

func (b *decodeBuffer) addStack(source exptrace.Stack) StackID {
	if source == exptrace.NoStack {
		return 0
	}
	if id := b.stackIDs[source]; id != 0 {
		return id
	}

	id := StackID(len(b.trace.Stacks) + 1)
	index := int(id - 1)
	if index == len(b.stackFrames) {
		b.stackFrames = append(b.stackFrames, nil)
	}
	frames := b.stackFrames[index][:0]
	for frame := range source.Frames() {
		frames = append(frames, Frame{
			PC:       frame.PC,
			Function: frame.Func,
			File:     frame.File,
			Line:     frame.Line,
		})
	}
	b.stackFrames[index] = frames
	b.stackCount = index + 1
	b.stackIDs[source] = id
	b.trace.Stacks[id] = Stack{Frames: frames}
	return id
}

func resource(source exptrace.ResourceID) Resource {
	result := Resource{Kind: ResourceNone}
	switch source.Kind {
	case exptrace.ResourceGoroutine:
		result.Kind, result.ID = ResourceGoroutine, int64(source.Goroutine())
	case exptrace.ResourceProc:
		result.Kind, result.ID = ResourceProcessor, int64(source.Proc())
	case exptrace.ResourceThread:
		result.Kind, result.ID = ResourceThread, int64(source.Thread())
	}
	return result
}

func states(transition exptrace.StateTransition) (State, State) {
	switch transition.Resource.Kind {
	case exptrace.ResourceGoroutine:
		from, to := transition.Goroutine()
		return goroutineState(from), goroutineState(to)
	case exptrace.ResourceProc:
		from, to := transition.Proc()
		return processorState(from), processorState(to)
	default:
		return StateUnknown, StateUnknown
	}
}

func goroutineState(state exptrace.GoState) State {
	switch state {
	case exptrace.GoNotExist:
		return StateNotExist
	case exptrace.GoRunnable:
		return StateRunnable
	case exptrace.GoRunning:
		return StateRunning
	case exptrace.GoWaiting:
		return StateWaiting
	case exptrace.GoSyscall:
		return StateSyscall
	default:
		return StateUnknown
	}
}

func processorState(state exptrace.ProcState) State {
	switch state {
	case exptrace.ProcNotExist:
		return StateNotExist
	case exptrace.ProcRunning:
		return StateRunning
	case exptrace.ProcIdle:
		return StateIdle
	default:
		return StateUnknown
	}
}
