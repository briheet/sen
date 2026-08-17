// Package runtime manages a Zig target and collects its runtime data.
package runtime

import (
	"bytes"
	"strconv"
	"sync"
	"time"

	"github.com/briheet/senbon/internal/model"
)

// frameOf resolves a runtime PC to a source frame, if possible.
type frameOf func(pc uint64) (file string, line uint64, ok bool)

// decodeWorkspace holds per-window scratch reused across decode calls.
type decodeWorkspace struct {
	lines       [][]byte
	fields      [][]byte
	stack       []model.ProfileLocationID
	frames      []model.TraceFrame
	key         []byte
	locationIDs map[uint64]model.ProfileLocationID
	stackIDs    map[string]model.StackID
}

var decodeWorkspaces = sync.Pool{New: func() any {
	return &decodeWorkspace{
		locationIDs: make(map[uint64]model.ProfileLocationID),
		stackIDs:    make(map[string]model.StackID),
	}
}}

// decodeSamples converts raw sample lines into the normalized profile and trace.
func decodeSamples(data []byte, interval time.Duration, frameOf frameOf) (*model.Profile, *model.Trace) {
	lineCount := bytes.Count(data, []byte{'\n'}) + 1
	profile := &model.Profile{
		SampleTypes: []model.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		PeriodType:  model.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:      int64(interval),
		Locations:   make(map[model.ProfileLocationID]model.ProfileLocation),
		Samples:     make([]model.ProfileSample, 0, lineCount),
	}
	trace := &model.Trace{
		Stacks: make(map[model.StackID]model.TraceStack),
		Events: make([]model.Event, 0, lineCount),
	}

	workspace := decodeWorkspaces.Get().(*decodeWorkspace)
	defer decodeWorkspaces.Put(workspace)
	clear(workspace.locationIDs)
	clear(workspace.stackIDs)

	lines := workspace.lines[:0]
	start := 0
	for index, byte := range data {
		if byte == '\n' {
			lines = append(lines, data[start:index])
			start = index + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	stack := workspace.stack[:0]
	frames := workspace.frames[:0]
	key := workspace.key[:0]
	values := []int64{int64(interval)} // shared by all samples: single-valued profile
	var elapsed time.Duration

	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		stack = stack[:0]
		frames = frames[:0]
		key = key[:0]
		fields := workspace.fields[:0]
		start = -1
		for index, byte := range line {
			if byte == ' ' || byte == '\t' || byte == '\r' {
				if start >= 0 {
					fields = append(fields, line[start:index])
					start = -1
				}
				continue
			}
			if start < 0 {
				start = index
			}
		}
		if start >= 0 {
			fields = append(fields, line[start:])
		}
		workspace.fields = fields

		for _, field := range fields {
			pc, ok := parseHex(field)
			if !ok {
				continue
			}
			id, ok := workspace.locationIDs[pc]
			if !ok {
				id = model.ProfileLocationID(len(profile.Locations) + 1)
				workspace.locationIDs[pc] = id
				location := model.ProfileLocation{ID: id}
				if file, line, ok := frameOf(pc); ok {
					location.Frames = []model.ProfileFrame{{Function: "?", File: file, Line: int64(line)}}
				}
				profile.Locations[id] = location
			}
			stack = append(stack, id)
			if location := profile.Locations[id]; len(location.Frames) > 0 {
				frame := location.Frames[0]
				frames = append(frames, model.TraceFrame{File: frame.File, Line: uint64(frame.Line)})
				key = append(key, frame.File...)
				key = append(key, ':')
				key = strconv.AppendInt(key, frame.Line, 10)
				key = append(key, ' ')
			}
		}
		if len(stack) == 0 {
			continue
		}
		profile.Samples = append(profile.Samples, model.ProfileSample{
			Values: values,
			Stack:  append([]model.ProfileLocationID(nil), stack...),
		})

		stackID, ok := workspace.stackIDs[string(key)]
		if !ok {
			stackID = model.StackID(len(trace.Stacks) + 1)
			trace.Stacks[stackID] = model.TraceStack{Frames: append([]model.TraceFrame(nil), frames...)}
			workspace.stackIDs[string(key)] = stackID
		}
		trace.Events = append(trace.Events, model.Event{
			At:    elapsed,
			Kind:  model.EventStackSample,
			Stack: stackID,
		})
		elapsed += interval
	}

	workspace.lines = lines
	workspace.stack = stack
	workspace.frames = frames
	workspace.key = key
	profile.Duration = elapsed
	trace.Duration = elapsed
	return profile, trace
}

func parseHex(b []byte) (uint64, bool) {
	if len(b) == 0 || len(b) > 16 {
		return 0, false
	}
	var value uint64
	for _, c := range b {
		var digit uint64
		switch {
		case c >= '0' && c <= '9':
			digit = uint64(c - '0')
		case c >= 'a' && c <= 'f':
			digit = uint64(c-'a') + 10
		case c >= 'A' && c <= 'F':
			digit = uint64(c-'A') + 10
		default:
			return 0, false
		}
		value = value<<4 | digit
	}
	return value, true
}
