package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeGraph(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n\ngo 1.25\n"), 0o600))

	mainPath := filepath.Join(root, "main.go")
	workerPath := filepath.Join(root, "worker.go")
	externalPath := filepath.Join(t.TempDir(), "external.go")
	static := &StaticGraph{
		Nodes: map[NodeID]*StaticNode{
			1: {ID: 1, Syntax: Syntax{File: 1, Start: Position{Line: 1}, End: Position{Line: 20}}},
			2: {ID: 2, Syntax: Syntax{File: 1, Start: Position{Line: 5, Column: 2}, End: Position{Line: 8}}},
			3: {ID: 3, Syntax: Syntax{File: 2, Start: Position{Line: 1}, End: Position{Line: 10}}},
			4: {ID: 4, Syntax: Syntax{File: 3, Start: Position{Line: 1}, End: Position{Line: 10}}},
		},
		Files: map[FileID]*StaticFile{
			1: {ID: 1, Path: mainPath, Package: 1, Functions: []NodeID{1, 2}},
			2: {ID: 2, Path: workerPath, Package: 1, Functions: []NodeID{3}},
			3: {ID: 3, Path: externalPath, Package: 2, Functions: []NodeID{4}},
		},
		Packages: map[PackageID]*Package{
			1: {Path: "example.com/app", Name: "main"},
			2: {Path: "example.com/external", Name: "external"},
		},
	}

	profile := &Profile{
		Duration: 10 * time.Millisecond,
		SampleTypes: []ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
		Locations: map[ProfileLocationID]ProfileLocation{
			1: {ID: 1, Frames: []ProfileFrame{{File: mainPath, Line: 6}, {File: mainPath, Line: 10}}},
			2: {ID: 2, Frames: []ProfileFrame{{File: "example.com/app/worker.go", Line: 5}}},
			3: {ID: 3, Frames: []ProfileFrame{{File: externalPath, Line: 5}}},
		},
		Samples: []ProfileSample{
			{Values: []int64{2, 100}, Stack: []ProfileLocationID{1, 2, 1}},
			{Values: []int64{1, 20}, Stack: []ProfileLocationID{3}},
		},
	}

	stack := TraceStack{Frames: []TraceFrame{
		{File: mainPath, Line: 6},
		{File: mainPath, Line: 10},
		{File: workerPath, Line: 5},
	}}
	observedTrace := &Trace{
		Duration: 40 * time.Nanosecond,
		Stacks: map[StackID]TraceStack{
			1: stack,
			2: {Frames: []TraceFrame{{File: externalPath, Line: 5}}},
		},
		Events: []Event{
			{At: 0, Kind: EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, ResourceStack: 1, From: StateNotExist, To: StateRunnable},
			{At: 0, Kind: EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: Resource{Kind: ResourceProcessor, ID: 0}, From: StateNotExist, To: StateRunning},
			{At: 2, Kind: EventRangeBegin, Goroutine: -1, Processor: -1, Thread: -1, Name: "GC", Resource: Resource{Kind: ResourceNone}},
			{At: 5, Kind: EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, ResourceStack: 1, From: StateRunnable, To: StateRunning},
			{At: 7, Kind: EventRangeEnd, Goroutine: -1, Processor: -1, Thread: -1, Name: "GC", Resource: Resource{Kind: ResourceNone}},
			{At: 10, Kind: EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Stack: 1, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, ResourceStack: 1, From: StateRunning, To: StateWaiting},
			{At: 12, Kind: EventStackSample, Goroutine: 1, Processor: 0, Thread: 10, Stack: 1},
			{At: 13, Kind: EventStackSample, Goroutine: 1, Processor: 0, Thread: 10, Stack: 2},
			{At: 15, Kind: EventMetric, Goroutine: -1, Processor: -1, Thread: -1, Name: "/gc/heap/goal:bytes", Value: 42},
			{At: 20, Kind: EventStateTransition, Goroutine: -1, Processor: -1, Thread: -1, Resource: Resource{Kind: ResourceProcessor, ID: 0}, From: StateRunning, To: StateIdle},
			{At: 30, Kind: EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, ResourceStack: 1, From: StateWaiting, To: StateRunnable},
			{At: 35, Kind: EventStateTransition, Goroutine: 1, Processor: 0, Thread: 10, Resource: Resource{Kind: ResourceGoroutine, ID: 1}, From: StateRunnable, To: StateNotExist},
		},
	}
	metrics := &RuntimeMetrics{Go: GoMetrics{UserCPU: 1.25}}
	profiles := map[string]*Profile{"cpu": profile}

	result := BuildRuntimeGraph("example.com/app", static)
	require.Len(t, result.Files, 2)
	require.Len(t, result.Nodes, 3)
	require.NotContains(t, result.Files, FileID(3))
	require.NotContains(t, result.Nodes, NodeID(4))
	require.Zero(t, result.Global.Process.Go.UserCPU)
	result.ApplyUpdate(result.BuildUpdate(metrics, profiles, observedTrace))
	require.Equal(t, 1.25, result.Global.Process.Go.UserCPU)

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
	traceRunnable := Metric{Source: TraceSource, Name: string(StateRunnable), Unit: "nanoseconds"}
	traceWaiting := Metric{Source: TraceSource, Name: string(StateWaiting), Unit: "nanoseconds"}
	require.Equal(t, Cost{Self: 1, Cumulative: 1}, result.Nodes[2].Metrics[traceSamples])
	require.Equal(t, int64(1), result.Unmapped[traceSamples])
	require.Equal(t, Cost{Self: 10, Cumulative: 10}, result.Nodes[2].Metrics[traceRunnable])
	require.Equal(t, Cost{Self: 20, Cumulative: 20}, result.Nodes[2].Metrics[traceWaiting])
	require.Equal(t, Cost{Cumulative: 20}, result.Nodes[3].Metrics[traceWaiting])
	require.Positive(t, result.NodeEdges[NodeEdge{From: 1, To: 2}])
	require.Positive(t, result.FileEdges[FileEdge{From: 2, To: 1}])

	summary := result.Global.Trace
	require.Equal(t, 40*time.Nanosecond, summary.Duration)
	require.Equal(t, uint64(1), summary.Goroutines)
	require.Zero(t, summary.LiveGoroutines)
	require.Equal(t, uint64(1), summary.PeakGoroutines)
	require.Equal(t, uint64(1), summary.Processors)
	require.Equal(t, uint64(1), summary.Threads)
	require.Equal(t, uint64(2), summary.StackSamples)
	require.Equal(t, 10*time.Nanosecond, summary.GoroutineStates[StateRunnable])
	require.Equal(t, 5*time.Nanosecond, summary.GoroutineStates[StateRunning])
	require.Equal(t, 20*time.Nanosecond, summary.GoroutineStates[StateWaiting])
	require.Equal(t, 20*time.Nanosecond, summary.ProcessorStates[StateRunning])
	require.Equal(t, 20*time.Nanosecond, summary.ProcessorStates[StateIdle])
	require.Equal(t, 5*time.Nanosecond, summary.Ranges["GC"])
	require.Equal(t, uint64(42), summary.Metrics["/gc/heap/goal:bytes"])

	aggregateSummary := newTraceSummary()
	aggregateSummary.Duration = 40 * time.Nanosecond
	aggregateSummary.Goroutines = 1
	aggregateSummary.PeakGoroutines = 1
	aggregateSummary.Processors = 1
	aggregateSummary.Threads = 1
	aggregateSummary.StackSamples = 2
	aggregateSummary.GoroutineStates[StateRunnable] = 10 * time.Nanosecond
	aggregateSummary.GoroutineStates[StateRunning] = 5 * time.Nanosecond
	aggregateSummary.GoroutineStates[StateWaiting] = 20 * time.Nanosecond
	aggregateSummary.ProcessorStates[StateRunning] = 20 * time.Nanosecond
	aggregateSummary.ProcessorStates[StateIdle] = 20 * time.Nanosecond
	aggregateSummary.Ranges["GC"] = 5 * time.Nanosecond
	aggregateSummary.Metrics["/gc/heap/goal:bytes"] = 42
	aggregatedTrace := &Trace{
		Duration: 40 * time.Nanosecond,
		Stacks:   observedTrace.Stacks,
		Aggregate: &TraceAggregate{
			Summary: aggregateSummary,
			Stacks: map[StackID]TraceStackAggregate{
				1: {
					Samples:  1,
					Runnable: TraceStackCost{Duration: 10 * time.Nanosecond, Occurrences: 2},
					Waiting:  TraceStackCost{Duration: 20 * time.Nanosecond, Occurrences: 1},
				},
				2: {Samples: 1},
			},
		},
	}
	aggregatedResult := BuildRuntimeGraph("example.com/app", static)
	aggregatedResult.ApplyUpdate(aggregatedResult.BuildUpdate(metrics, profiles, aggregatedTrace))
	require.Equal(t, result.Global, aggregatedResult.Global)
	require.Equal(t, result.Nodes, aggregatedResult.Nodes)
	require.Equal(t, result.Files, aggregatedResult.Files)
	require.Equal(t, result.Unmapped, aggregatedResult.Unmapped)
	require.Equal(t, result.NodeEdges, aggregatedResult.NodeEdges)
	require.Equal(t, result.FileEdges, aggregatedResult.FileEdges)

	node := result.Nodes[2]
	result.ApplyUpdate(result.BuildUpdate(&RuntimeMetrics{Go: GoMetrics{UserCPU: 2.5}}, nil, nil))
	require.Same(t, node, result.Nodes[2])
	require.Empty(t, result.Nodes[2].Metrics)
	require.Empty(t, result.Global.ProfileTotals)
	require.Empty(t, result.Unmapped)
	require.Zero(t, result.Global.Trace.Duration)
	require.Equal(t, 2.5, result.Global.Process.Go.UserCPU)
}

func TestTraceEdgesFollowCallerToCalleeOrder(t *testing.T) {
	nodes := make(map[NodeEdge]int64)
	files := make(map[FileEdge]int64)
	addNodeTraceEdges(nodes, []NodeID{3, 2, 1})
	addFileTraceEdges(files, []FileID{2, 1})

	require.Equal(t, map[NodeEdge]int64{
		{From: 1, To: 2}: 1,
		{From: 2, To: 3}: 1,
	}, nodes)
	require.Equal(t, map[FileEdge]int64{{From: 1, To: 2}: 1}, files)
}

func TestTraceTargetsCacheStackAndFileResolution(t *testing.T) {
	const sourcePath = "/src/example/main.go"
	graph := BuildRuntimeGraph("example.com/app", &StaticGraph{
		Nodes: map[NodeID]*StaticNode{
			1: {ID: 1, Syntax: Syntax{File: 1, Start: Position{Line: 1}, End: Position{Line: 20}}},
		},
		Files: map[FileID]*StaticFile{
			1: {ID: 1, Path: sourcePath, Package: 1, Functions: []NodeID{1}},
		},
		Packages: map[PackageID]*Package{
			1: {Path: "example.com/app", Name: "main"},
		},
	})
	trace := &Trace{Stacks: map[StackID]TraceStack{
		1: {Frames: []TraceFrame{{File: "example.com/app/main.go", Line: 10}}},
	}}
	workspace := acquireTraceWorkspace()
	targets := acquireTargetWorkspace()
	t.Cleanup(func() {
		releaseTraceWorkspace(workspace)
		releaseTargetWorkspace(targets)
	})

	nodes, files := graph.traceTargets(trace, 1, workspace, targets)
	cachedNodes, cachedFiles := graph.traceTargets(trace, 1, workspace, targets)
	require.Equal(t, []NodeID{1}, nodes)
	require.Equal(t, []FileID{1}, files)
	require.Same(t, &nodes[0], &cachedNodes[0])
	require.Same(t, &files[0], &cachedFiles[0])
	require.Len(t, workspace.targets, 1)
	require.Equal(t, FileID(1), graph.mapper.resolved["example.com/app/main.go"])

	_, ok := graph.mapper.file("example.com/other/missing.go")
	require.False(t, ok)
	require.Contains(t, graph.mapper.resolved, "example.com/other/missing.go")
}

func TestSnapshotIncludesProfileActivity(t *testing.T) {
	graph := &RuntimeGraph{
		Nodes: map[NodeID]*Node{
			1: {Metrics: CodeMetrics{
				Metric{Source: "redis", Name: "calls", Unit: "count"}: {Self: 3, Cumulative: 3},
			}},
		},
		Files:     make(map[FileID]*File),
		NodeEdges: make(map[NodeEdge]int64),
		FileEdges: make(map[FileEdge]int64),
	}

	require.Equal(t, int64(1), graph.Snapshot().NodeActivity[1])
}
