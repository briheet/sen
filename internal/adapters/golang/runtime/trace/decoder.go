package trace

import (
	"io"

	exptrace "golang.org/x/exp/trace"
)

// Read decodes a Go runtime trace into sen's representation.
func Read(r io.Reader) (*Trace, error) {
	reader, err := exptrace.NewReader(r)
	if err != nil {
		return nil, err
	}

	result := &Trace{Stacks: make(map[StackID]Stack)}
	stackIDs := make(map[exptrace.Stack]StackID)
	var start exptrace.Time

	for {
		source, err := reader.ReadEvent()
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return nil, err
		}
		if source.Kind() == exptrace.EventExperimental {
			continue
		}
		if len(result.Events) == 0 {
			start = source.Time()
		}

		event := Event{
			At:        source.Time().Sub(start),
			Kind:      eventKind(source.Kind()),
			Goroutine: int64(source.Goroutine()),
			Processor: int64(source.Proc()),
			Thread:    int64(source.Thread()),
			Stack:     addStack(source.Stack(), stackIDs, result.Stacks),
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
			event.ResourceStack = addStack(transition.Stack, stackIDs, result.Stacks)
			event.Reason = transition.Reason
			event.From, event.To = states(transition)
		}

		result.Events = append(result.Events, event)
		result.Duration = event.At
	}
}

func addStack(source exptrace.Stack, ids map[exptrace.Stack]StackID, stacks map[StackID]Stack) StackID {
	if source == exptrace.NoStack {
		return 0
	}
	if id := ids[source]; id != 0 {
		return id
	}

	id := StackID(len(stacks) + 1)
	stack := Stack{}
	for frame := range source.Frames() {
		stack.Frames = append(stack.Frames, Frame{
			PC:       frame.PC,
			Function: frame.Func,
			File:     frame.File,
			Line:     frame.Line,
		})
	}
	ids[source] = id
	stacks[id] = stack
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
