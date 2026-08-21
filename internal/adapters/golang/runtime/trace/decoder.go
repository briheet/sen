package trace

import (
	"io"

	exptrace "golang.org/x/exp/trace"
)

const (
	maxReusableEvents = 64 * 1024
	maxReusableStacks = 4 * 1024
	maxReusableFrames = 1024
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
	if cap(b.trace.Events) > maxReusableEvents {
		b.trace.Events = nil
	} else {
		clear(b.trace.Events)
		b.trace.Events = b.trace.Events[:0]
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

	var start exptrace.Time

	for {
		source, err := reader.ReadEvent()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if source.Kind() == exptrace.EventExperimental {
			continue
		}
		if len(b.trace.Events) == 0 {
			start = source.Time()
		}

		event := Event{
			At:        source.Time().Sub(start),
			Kind:      eventKind(source.Kind()),
			Goroutine: int64(source.Goroutine()),
			Processor: int64(source.Proc()),
			Thread:    int64(source.Thread()),
			Stack:     b.addStack(source.Stack()),
		}

		switch source.Kind() {
		case exptrace.EventMetric:
			metric := source.Metric()
			event.Name = metric.Name
			event.Value = metric.Value.Uint64()
		case exptrace.EventLabel:
			label := source.Label()
			event.Name = label.Label
			event.Resource = resource(label.Resource)
		case exptrace.EventRangeBegin, exptrace.EventRangeActive, exptrace.EventRangeEnd:
			rangeEvent := source.Range()
			event.Name = rangeEvent.Name
			event.Resource = resource(rangeEvent.Scope)
		case exptrace.EventTaskBegin, exptrace.EventTaskEnd:
			task := source.Task()
			event.Task = uint64(task.ID)
			event.Parent = uint64(task.Parent)
			event.Name = task.Type
		case exptrace.EventRegionBegin, exptrace.EventRegionEnd:
			region := source.Region()
			event.Task = uint64(region.Task)
			event.Name = region.Type
		case exptrace.EventLog:
			log := source.Log()
			event.Task = uint64(log.Task)
			event.Category = log.Category
			event.Message = log.Message
		case exptrace.EventStateTransition:
			transition := source.StateTransition()
			event.Resource = resource(transition.Resource)
			event.ResourceStack = b.addStack(transition.Stack)
			event.Reason = transition.Reason
			event.From, event.To = states(transition)
		}

		b.trace.Events = append(b.trace.Events, event)
		b.trace.Duration = event.At
	}
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

func eventKind(kind exptrace.EventKind) EventKind {
	switch kind {
	case exptrace.EventSync:
		return EventSync
	case exptrace.EventMetric:
		return EventMetric
	case exptrace.EventLabel:
		return EventLabel
	case exptrace.EventStackSample:
		return EventStackSample
	case exptrace.EventRangeBegin:
		return EventRangeBegin
	case exptrace.EventRangeActive:
		return EventRangeActive
	case exptrace.EventRangeEnd:
		return EventRangeEnd
	case exptrace.EventTaskBegin:
		return EventTaskBegin
	case exptrace.EventTaskEnd:
		return EventTaskEnd
	case exptrace.EventRegionBegin:
		return EventRegionBegin
	case exptrace.EventRegionEnd:
		return EventRegionEnd
	case exptrace.EventLog:
		return EventLog
	case exptrace.EventStateTransition:
		return EventStateTransition
	default:
		return "unknown"
	}
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
