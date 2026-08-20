// Package profiler captures native Rust CPU stacks on supported host platforms.
package profiler

import (
	"strconv"
	"strings"
	"time"

	"github.com/briheet/sen/internal/adapters/rust/analysis"
	"github.com/briheet/sen/internal/model"
)

const (
	window = time.Second
	period = 10 * time.Millisecond
)

type rawFrame struct {
	address  uint64
	function string
	file     string
	line     int64
}

type rawSample struct {
	frames []rawFrame // leaf first
	weight uint64
}

func normalize(samples []rawSample, symbols *analysis.Symbols) (*model.Profile, *model.Trace, uint64) {
	profile := &model.Profile{
		Duration:          window,
		SampleTypes:       []model.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		DefaultSampleType: "cpu",
		PeriodType:        model.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:            int64(period),
		Locations:         make(map[model.ProfileLocationID]model.ProfileLocation),
	}
	trace := &model.Trace{Duration: window, Stacks: make(map[model.StackID]model.TraceStack)}
	locations := make(map[string]model.ProfileLocationID)
	stacks := make(map[string]model.StackID)
	var elapsed time.Duration
	var total uint64
	for _, sample := range samples {
		if sample.weight == 0 {
			sample.weight = 1
		}
		ids := make([]model.ProfileLocationID, 0, len(sample.frames))
		frames := make([]model.TraceFrame, 0, len(sample.frames))
		for _, raw := range sample.frames {
			resolved := symbols.Symbolize(raw.address, raw.function)
			if raw.file != "" {
				resolved.File = raw.file
			}
			if raw.line > 0 {
				resolved.Line = raw.line
			}
			key := resolved.Function + "\x00" + resolved.File + "\x00" + strconv.FormatInt(resolved.Line, 10)
			id, ok := locations[key]
			if !ok {
				id = model.ProfileLocationID(len(locations) + 1)
				locations[key] = id
				profile.Locations[id] = model.ProfileLocation{ID: id, Address: raw.address, Frames: []model.ProfileFrame{{Function: resolved.Function, File: resolved.File, Line: resolved.Line}}}
			}
			ids = append(ids, id)
			frames = append(frames, model.TraceFrame{Function: resolved.Function, File: resolved.File, Line: uint64(max(0, resolved.Line))})
		}
		if len(ids) == 0 {
			continue
		}
		profile.Samples = append(profile.Samples, model.ProfileSample{Values: []int64{int64(sample.weight) * int64(period)}, Stack: ids})
		stackKey := stackKey(ids)
		stackID, ok := stacks[stackKey]
		if !ok {
			stackID = model.StackID(len(stacks) + 1)
			stacks[stackKey] = stackID
			trace.Stacks[stackID] = model.TraceStack{Frames: frames}
		}
		for range sample.weight {
			elapsed += period
			trace.Events = append(trace.Events, model.Event{At: elapsed, Kind: model.EventStackSample, Stack: stackID})
		}
		total += sample.weight
	}
	return profile, trace, total
}

func stackKey(ids []model.ProfileLocationID) string {
	var builder strings.Builder
	for _, id := range ids {
		builder.WriteString(strconv.FormatUint(uint64(id), 10))
		builder.WriteByte('/')
	}
	return builder.String()
}

func sourceLocation(value string) (string, int64) {
	value = strings.Trim(strings.TrimSpace(value), "()[]")
	index := strings.LastIndexByte(value, ':')
	if index <= 0 {
		return "", 0
	}
	line, err := strconv.ParseInt(value[index+1:], 10, 64)
	if err != nil {
		return "", 0
	}
	return value[:index], line
}
