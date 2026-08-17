package model

import (
	stdruntimemetrics "runtime/metrics"
	"testing"
	"time"

	graph "github.com/briheet/senbon/internal/analysis"
	"github.com/briheet/senbon/internal/runtime"
	runtimemetrics "github.com/briheet/senbon/internal/runtime/metrics"
	runtimepprof "github.com/briheet/senbon/internal/runtime/pprof"
	runtimetrace "github.com/briheet/senbon/internal/runtime/trace"
)

func BenchmarkRuntimeGraphUpdates(b *testing.B) {
	const sourcePath = "/src/example/main.go"
	static := &graph.Graph{
		Nodes: map[graph.NodeID]*graph.Node{
			1: {ID: 1, Syntax: graph.Syntax{File: 1, Start: graph.Position{Line: 1}, End: graph.Position{Line: 20}}},
		},
		Files: map[graph.FileID]*graph.File{
			1: {ID: 1, Path: sourcePath, Package: 1, Functions: []graph.NodeID{1}},
		},
		Packages: map[graph.PackageID]*graph.Package{
			1: {Path: "example.com/app", Name: "main"},
		},
	}
	profile := &runtimepprof.Profile{
		Duration:    time.Second,
		SampleTypes: []runtimepprof.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Locations: map[runtimepprof.LocationID]runtimepprof.Location{
			1: {ID: 1, Frames: []runtimepprof.Frame{{File: sourcePath, Line: 10}}},
		},
		Samples: []runtimepprof.Sample{{Values: []int64{100}, Stack: []runtimepprof.LocationID{1}}},
	}
	trace := &runtimetrace.Trace{
		Duration: time.Second,
		Stacks: map[runtimetrace.StackID]runtimetrace.Stack{
			1: {Frames: []runtimetrace.Frame{{File: sourcePath, Line: 10}}},
		},
		Events: []runtimetrace.Event{
			{Kind: runtimetrace.EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, ResourceStack: 1, From: runtimetrace.StateNotExist, To: runtimetrace.StateRunnable},
			{At: time.Second, Kind: runtimetrace.EventStateTransition, Goroutine: 1, Processor: 0, Thread: 1, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, From: runtimetrace.StateRunnable, To: runtimetrace.StateNotExist},
		},
	}
	observed := &runtime.Runtime{
		Metrics: &runtimemetrics.RuntimeMetrics{
			UserCPU:          1,
			SchedulerLatency: &stdruntimemetrics.Float64Histogram{Counts: []uint64{1}, Buckets: []float64{0, 1}},
			GCPauses:         &stdruntimemetrics.Float64Histogram{Counts: []uint64{1}, Buckets: []float64{0, 1}},
		},
		Profiles: map[string]*runtimepprof.Profile{"cpu": profile},
		Trace:    trace,
	}
	result := BuildRuntimeGraph("example.com/app", static)

	b.ReportAllocs()
	for b.Loop() {
		result.ApplyUpdate(result.BuildUpdate(observed))
	}
}
