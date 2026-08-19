package model

import (
	"testing"
	"time"
)

func BenchmarkRuntimeGraphUpdates(b *testing.B) {
	const sourcePath = "/src/example/main.go"
	static := &StaticGraph{
		Nodes: map[NodeID]*StaticNode{
			1: {ID: 1, Syntax: Syntax{File: 1, Start: Position{Line: 1}, End: Position{Line: 20}}},
		},
		Files: map[FileID]*StaticFile{
			1: {ID: 1, Path: sourcePath, Package: 1, Functions: []NodeID{1}},
		},
		Packages: map[PackageID]*Package{
			1: {Path: "example.com/app", Name: "main"},
		},
	}
	profile := &Profile{
		Duration:    time.Second,
		SampleTypes: []ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Locations: map[ProfileLocationID]ProfileLocation{
			1: {ID: 1, Frames: []ProfileFrame{{File: sourcePath, Line: 10}}},
		},
		Samples: []ProfileSample{{Values: []int64{100}, Stack: []ProfileLocationID{1}}},
	}
	trace := &Trace{
		Duration: time.Second,
		Stacks: map[StackID]TraceStack{
			1: {Frames: []TraceFrame{{File: sourcePath, Line: 10}}},
		},
		Events: []Event{
			{Kind: EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, ResourceStack: 1, From: StateNotExist, To: StateRunnable},
			{At: time.Second, Kind: EventStateTransition, Goroutine: 1, Processor: 0, Thread: 1, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, From: StateRunnable, To: StateNotExist},
		},
	}
	metrics := &RuntimeMetrics{
		Go: GoMetrics{
			UserCPU:          1,
			SchedulerLatency: &Histogram{Counts: []uint64{1}, Buckets: []float64{0, 1}},
			GCPauses:         &Histogram{Counts: []uint64{1}, Buckets: []float64{0, 1}},
		},
	}
	profiles := map[string]*Profile{"cpu": profile}
	result := BuildRuntimeGraph("example.com/app", static)

	b.ReportAllocs()
	for b.Loop() {
		result.ApplyUpdate(result.BuildUpdate(metrics, profiles, trace))
	}
}
