package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	graph "github.com/briheet/senbon/internal/analysis"
	"github.com/briheet/senbon/internal/runtime"
	runtimemetrics "github.com/briheet/senbon/internal/runtime/metrics"
	runtimepprof "github.com/briheet/senbon/internal/runtime/pprof"
	runtimetrace "github.com/briheet/senbon/internal/runtime/trace"
	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeGraph(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n\ngo 1.25\n"), 0o600))

	mainPath := filepath.Join(root, "main.go")
	workerPath := filepath.Join(root, "worker.go")
	externalPath := filepath.Join(t.TempDir(), "external.go")
	static := &graph.Graph{
		Nodes: map[graph.NodeID]*graph.Node{
			1: {ID: 1, Syntax: graph.Syntax{File: 1, Start: graph.Position{Line: 1}, End: graph.Position{Line: 20}}},
			2: {ID: 2, Syntax: graph.Syntax{File: 1, Start: graph.Position{Line: 5, Column: 2}, End: graph.Position{Line: 8}}},
			3: {ID: 3, Syntax: graph.Syntax{File: 2, Start: graph.Position{Line: 1}, End: graph.Position{Line: 10}}},
			4: {ID: 4, Syntax: graph.Syntax{File: 3, Start: graph.Position{Line: 1}, End: graph.Position{Line: 10}}},
		},
		Files: map[graph.FileID]*graph.File{
			1: {ID: 1, Path: mainPath, Package: 1, Functions: []graph.NodeID{1, 2}},
			2: {ID: 2, Path: workerPath, Package: 1, Functions: []graph.NodeID{3}},
			3: {ID: 3, Path: externalPath, Package: 2, Functions: []graph.NodeID{4}},
		},
		Packages: map[graph.PackageID]*graph.Package{
			1: {Path: "example.com/app", Name: "main"},
			2: {Path: "example.com/external", Name: "external"},
		},
	}

	profile := &runtimepprof.Profile{
		Duration: 10 * time.Millisecond,
		SampleTypes: []runtimepprof.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
		Locations: map[runtimepprof.LocationID]runtimepprof.Location{
			1: {ID: 1, Frames: []runtimepprof.Frame{{File: mainPath, Line: 6}, {File: mainPath, Line: 10}}},
			2: {ID: 2, Frames: []runtimepprof.Frame{{File: "example.com/app/worker.go", Line: 5}}},
			3: {ID: 3, Frames: []runtimepprof.Frame{{File: externalPath, Line: 5}}},
		},
		Samples: []runtimepprof.Sample{
			{Values: []int64{2, 100}, Stack: []runtimepprof.LocationID{1, 2, 1}},
			{Values: []int64{1, 20}, Stack: []runtimepprof.LocationID{3}},
		},
	}

	stack := runtimetrace.Stack{Frames: []runtimetrace.Frame{
		{File: mainPath, Line: 6},
		{File: mainPath, Line: 10},
		{File: workerPath, Line: 5},
	}}
	observedTrace := &runtimetrace.Trace{
		Duration: 40 * time.Nanosecond,
		Stacks: map[runtimetrace.StackID]runtimetrace.Stack{
			1: stack,
			2: {Frames: []runtimetrace.Frame{{File: externalPath, Line: 5}}},
		},
		Events: []runtimetrace.Event{
			{At: 0, Kind: runtimetrace.EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, ResourceStack: 1, From: runtimetrace.StateNotExist, To: runtimetrace.StateRunnable},
			{At: 0, Kind: runtimetrace.EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceProcessor, ID: 0}, From: runtimetrace.StateNotExist, To: runtimetrace.StateRunning},
			{At: 2, Kind: runtimetrace.EventRangeBegin, Goroutine: -1, Processor: -1, Thread: -1, Name: "GC", Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceNone}},
			{At: 5, Kind: runtimetrace.EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, ResourceStack: 1, From: runtimetrace.StateRunnable, To: runtimetrace.StateRunning},
			{At: 7, Kind: runtimetrace.EventRangeEnd, Goroutine: -1, Processor: -1, Thread: -1, Name: "GC", Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceNone}},
			{At: 10, Kind: runtimetrace.EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Stack: 1, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, ResourceStack: 1, From: runtimetrace.StateRunning, To: runtimetrace.StateWaiting},
			{At: 12, Kind: runtimetrace.EventStackSample, Goroutine: 1, Processor: 0, Thread: 10, Stack: 1},
			{At: 13, Kind: runtimetrace.EventStackSample, Goroutine: 1, Processor: 0, Thread: 10, Stack: 2},
			{At: 15, Kind: runtimetrace.EventMetric, Goroutine: -1, Processor: -1, Thread: -1, Name: "/gc/heap/goal:bytes", Value: 42},
			{At: 20, Kind: runtimetrace.EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceProcessor, ID: 0}, From: runtimetrace.StateRunning, To: runtimetrace.StateIdle},
			{At: 30, Kind: runtimetrace.EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, ResourceStack: 1, From: runtimetrace.StateWaiting, To: runtimetrace.StateRunnable},
			{At: 35, Kind: runtimetrace.EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Resource: runtimetrace.Resource{Kind: runtimetrace.ResourceGoroutine, ID: 1}, From: runtimetrace.StateRunnable, To: runtimetrace.StateNotExist},
		},
	}
	observed := &runtime.Runtime{
		Metrics:  &runtimemetrics.RuntimeMetrics{UserCPU: 1.25},
		Profiles: map[string]*runtimepprof.Profile{"cpu": profile},
		Trace:    observedTrace,
	}

	result := BuildRuntimeGraph("example.com/app", static)
	require.Len(t, result.Files, 2)
	require.Len(t, result.Nodes, 3)
	require.NotContains(t, result.Files, graph.FileID(3))
	require.NotContains(t, result.Nodes, graph.NodeID(4))
	require.Zero(t, result.Global.Process.UserCPU)
	result.ApplyUpdate(result.BuildUpdate(observed))
	require.Equal(t, 1.25, result.Global.Process.UserCPU)

	profileCount := Metric{Source: "cpu", Name: "samples", Unit: "count"}
	profileCPU := Metric{Source: "cpu", Name: "cpu", Unit: "nanoseconds"}
	require.Equal(t, int64(3), result.Global.ProfileTotals[profileCount])
	require.Equal(t, int64(120), result.Global.ProfileTotals[profileCPU])
	require.Equal(t, 10*time.Millisecond, result.Global.ProfileDurations["cpu"])
	require.Equal(t, Cost{Self: 2, Cumulative: 2}, result.Nodes[2].Metrics[profileCount])
	require.Equal(t, Cost{Cumulative: 2}, result.Nodes[1].Metrics[profileCount])
	require.Equal(t, Cost{Cumulative: 2}, result.Nodes[3].Metrics[profileCount])
	require.Equal(t, Cost{Self: 2, Cumulative: 2}, result.Files[1].Metrics[profileCount])
	require.Equal(t, Cost{Cumulative: 2}, result.Files[2].Metrics[profileCount])
	require.Equal(t, int64(1), result.Unmapped[profileCount])

	traceSamples := Metric{Source: TraceSource, Name: "samples", Unit: "count"}
	traceRunnable := Metric{Source: TraceSource, Name: string(runtimetrace.StateRunnable), Unit: "nanoseconds"}
	traceWaiting := Metric{Source: TraceSource, Name: string(runtimetrace.StateWaiting), Unit: "nanoseconds"}
	require.Equal(t, Cost{Self: 1, Cumulative: 1}, result.Nodes[2].Metrics[traceSamples])
	require.Equal(t, int64(1), result.Unmapped[traceSamples])
	require.Equal(t, Cost{Self: 10, Cumulative: 10}, result.Nodes[2].Metrics[traceRunnable])
	require.Equal(t, Cost{Self: 20, Cumulative: 20}, result.Nodes[2].Metrics[traceWaiting])
	require.Equal(t, Cost{Cumulative: 20}, result.Nodes[3].Metrics[traceWaiting])

	summary := result.Global.Trace
	require.Equal(t, 40*time.Nanosecond, summary.Duration)
	require.Equal(t, uint64(1), summary.Goroutines)
	require.Zero(t, summary.LiveGoroutines)
	require.Equal(t, uint64(1), summary.PeakGoroutines)
	require.Equal(t, uint64(1), summary.Processors)
	require.Equal(t, uint64(1), summary.Threads)
	require.Equal(t, uint64(2), summary.StackSamples)
	require.Equal(t, 10*time.Nanosecond, summary.GoroutineStates[runtimetrace.StateRunnable])
	require.Equal(t, 5*time.Nanosecond, summary.GoroutineStates[runtimetrace.StateRunning])
	require.Equal(t, 20*time.Nanosecond, summary.GoroutineStates[runtimetrace.StateWaiting])
	require.Equal(t, 20*time.Nanosecond, summary.ProcessorStates[runtimetrace.StateRunning])
	require.Equal(t, 20*time.Nanosecond, summary.ProcessorStates[runtimetrace.StateIdle])
	require.Equal(t, 5*time.Nanosecond, summary.Ranges["GC"])
	require.Equal(t, uint64(42), summary.Metrics["/gc/heap/goal:bytes"])

	node := result.Nodes[2]
	result.ApplyUpdate(result.BuildMetricsUpdate(&runtimemetrics.RuntimeMetrics{UserCPU: 2}))
	require.Equal(t, 2.0, result.Global.Process.UserCPU)
	require.Equal(t, int64(3), result.Global.ProfileTotals[profileCount])
	require.Equal(t, 40*time.Nanosecond, result.Global.Trace.Duration)

	result.ApplyUpdate(result.BuildProfileUpdate("cpu", nil))
	require.Empty(t, result.Global.ProfileTotals)
	require.Empty(t, result.Global.ProfileDurations)
	require.Equal(t, 40*time.Nanosecond, result.Global.Trace.Duration)

	result.ApplyUpdate(result.BuildUpdate(&runtime.Runtime{Metrics: &runtimemetrics.RuntimeMetrics{UserCPU: 2.5}}))
	require.Same(t, node, result.Nodes[2])
	require.Empty(t, result.Nodes[2].Metrics)
	require.Empty(t, result.Global.ProfileTotals)
	require.Empty(t, result.Unmapped)
	require.Zero(t, result.Global.Trace.Duration)
	require.Equal(t, 2.5, result.Global.Process.UserCPU)
}
