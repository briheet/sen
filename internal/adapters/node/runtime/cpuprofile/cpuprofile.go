// Package cpuprofile decodes V8 CPU profiles into sen's model.
package cpuprofile

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/briheet/sen/internal/model"
)

const (
	sampleIntervalMicros = 1000
	maxPooledMapEntries  = 4096
)

// decodeWorkspace holds per-window scratch reused across profile windows.
type decodeWorkspace struct {
	parents   map[uint32]uint32
	frames    map[uint32]CallFrame
	locations map[uint32][]model.ProfileLocationID
	traceIDs  map[uint32]model.StackID
}

var decodeWorkspaces = sync.Pool{New: func() any {
	return &decodeWorkspace{
		parents:   make(map[uint32]uint32),
		frames:    make(map[uint32]CallFrame),
		locations: make(map[uint32][]model.ProfileLocationID),
		traceIDs:  make(map[uint32]model.StackID),
	}
}}

// CPUProfile mirrors the V8/Chrome CPU profile JSON format.
type CPUProfile struct {
	Nodes      []Node   `json:"nodes"`
	Samples    []uint32 `json:"samples"`
	TimeDeltas []int64  `json:"timeDeltas"`
	StartTime  int64    `json:"startTime"`
	EndTime    int64    `json:"endTime"`
}

// Node is one entry in the profile's flattened call tree.
type Node struct {
	ID        uint32    `json:"id"`
	CallFrame CallFrame `json:"callFrame"`
	Children  []uint32  `json:"children"`
}

// CallFrame describes one stack frame.
type CallFrame struct {
	FunctionName string `json:"functionName"`
	URL          string `json:"url"`
	LineNumber   int    `json:"lineNumber"`
	ColumnNumber int    `json:"columnNumber"`
}

// Profile converts the CPU profile into the normalized profile model.
func (p *CPUProfile) Profile() *model.Profile {
	workspace := acquireDecodeWorkspace()
	defer releaseDecodeWorkspace(workspace)
	p.fillParents(workspace.parents)

	result := &model.Profile{
		Duration:    time.Duration(p.EndTime-p.StartTime) * time.Microsecond,
		SampleTypes: []model.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		PeriodType:  model.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:      sampleIntervalMicros * 1000,
		Locations:   make(map[model.ProfileLocationID]model.ProfileLocation, len(p.Nodes)),
		Samples:     make([]model.ProfileSample, 0, len(p.Samples)),
	}
	for _, node := range p.Nodes {
		result.Locations[model.ProfileLocationID(node.ID)] = model.ProfileLocation{
			ID: model.ProfileLocationID(node.ID),
			Frames: []model.ProfileFrame{{
				Function: node.CallFrame.FunctionName,
				File:     file(node.CallFrame.URL),
				Line:     int64(node.CallFrame.LineNumber + 1),
				Column:   int64(node.CallFrame.ColumnNumber + 1),
			}},
		}
	}
	for index, sample := range p.Samples {
		if index >= len(p.TimeDeltas) {
			break
		}
		stack, ok := workspace.locations[sample]
		if !ok {
			stack = p.stack(sample, workspace.parents)
			workspace.locations[sample] = stack
		}
		result.Samples = append(result.Samples, model.ProfileSample{
			Values: []int64{p.TimeDeltas[index] * 1000},
			Stack:  stack,
		})
	}
	return result
}

// Trace converts the profile's sample timeline into the normalized trace model.
func (p *CPUProfile) Trace() *model.Trace {
	workspace := acquireDecodeWorkspace()
	defer releaseDecodeWorkspace(workspace)
	p.fillParents(workspace.parents)
	p.fillFrames(workspace.frames)

	trace := &model.Trace{
		Duration: time.Duration(p.EndTime-p.StartTime) * time.Microsecond,
		Stacks:   make(map[model.StackID]model.TraceStack),
		Events:   make([]model.Event, 0, len(p.Samples)),
	}
	var elapsed time.Duration
	for index, sample := range p.Samples {
		if index >= len(p.TimeDeltas) {
			break
		}
		elapsed += time.Duration(p.TimeDeltas[index]) * time.Microsecond
		stackID, ok := workspace.traceIDs[sample]
		if !ok {
			stackID = registerStack(trace, p.stackFrames(sample, workspace.parents, workspace.frames))
			workspace.traceIDs[sample] = stackID
		}
		trace.Events = append(trace.Events, model.Event{
			At:    elapsed,
			Kind:  model.EventStackSample,
			Stack: stackID,
		})
	}
	return trace
}

func acquireDecodeWorkspace() *decodeWorkspace {
	return decodeWorkspaces.Get().(*decodeWorkspace)
}

func releaseDecodeWorkspace(workspace *decodeWorkspace) {
	resetDecodeMap(&workspace.parents)
	resetDecodeMap(&workspace.frames)
	resetDecodeMap(&workspace.locations)
	resetDecodeMap(&workspace.traceIDs)
	decodeWorkspaces.Put(workspace)
}

func resetDecodeMap[K comparable, V any](target *map[K]V) {
	if len(*target) > maxPooledMapEntries {
		*target = make(map[K]V)
		return
	}
	clear(*target)
}

// registerStack stores a new stack, returning zero if it has no frames.
func registerStack(trace *model.Trace, frames []model.TraceFrame) model.StackID {
	if len(frames) == 0 {
		return 0
	}
	id := model.StackID(len(trace.Stacks) + 1)
	trace.Stacks[id] = model.TraceStack{Frames: frames}
	return id
}

// fillParents inverts the children lists into the caller's map.
func (p *CPUProfile) fillParents(parents map[uint32]uint32) {
	for _, node := range p.Nodes {
		for _, child := range node.Children {
			parents[child] = node.ID
		}
	}
}

// fillFrames indexes call frames by node id into the caller's map.
func (p *CPUProfile) fillFrames(frames map[uint32]CallFrame) {
	for _, node := range p.Nodes {
		frames[node.ID] = node.CallFrame
	}
}

// stack returns a leaf-first location chain for a sample.
func (p *CPUProfile) stack(leaf uint32, parents map[uint32]uint32) []model.ProfileLocationID {
	var stack []model.ProfileLocationID
	for id := leaf; id != 0; id = parents[id] {
		stack = append(stack, model.ProfileLocationID(id))
	}
	return stack
}

// stackFrames returns a leaf-first frame chain for a sample.
func (p *CPUProfile) stackFrames(leaf uint32, parents map[uint32]uint32, frames map[uint32]CallFrame) []model.TraceFrame {
	var stack []model.TraceFrame
	for id := leaf; id != 0; id = parents[id] {
		frame, ok := frames[id]
		if !ok {
			break
		}
		stack = append(stack, model.TraceFrame{
			Function: frame.FunctionName,
			File:     file(frame.URL),
			Line:     uint64(frame.LineNumber + 1),
		})
	}
	return stack
}

// file normalizes a V8 script URL into an absolute source path.
func file(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	rawURL, _ = strings.CutPrefix(rawURL, "file://")
	return filepath.Clean(rawURL)
}
