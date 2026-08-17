// Package model merges static analysis with observed runtime data.
package model

import (
	"maps"
	"time"

	graph "github.com/briheet/senbon/internal/analysis"
	"github.com/briheet/senbon/internal/runtime"
	runtimemetrics "github.com/briheet/senbon/internal/runtime/metrics"
	runtimepprof "github.com/briheet/senbon/internal/runtime/pprof"
	runtimetrace "github.com/briheet/senbon/internal/runtime/trace"
)

const (
	// TraceSource identifies measurements derived from runtime traces.
	TraceSource     = "trace"
	traceSamples    = "samples"
	unitCount       = "count"
	unitNanoseconds = "nanoseconds"
)

// RuntimeGraph is the TUI-owned view of static and observed program data.
type RuntimeGraph struct {
	Static   *graph.Graph
	Nodes    map[graph.NodeID]*Node
	Files    map[graph.FileID]*File
	Global   Global
	Unmapped map[Metric]int64

	mapper *frameMapper
}

// Node combines a static function node with its runtime costs.
type Node struct {
	Static  *graph.Node
	Metrics CodeMetrics
}

// File combines a static source file with its runtime costs.
type File struct {
	Static  *graph.File
	Metrics CodeMetrics
}

// Metric identifies one observed measurement.
type Metric struct {
	Source string
	Name   string
	Unit   string
}

// Cost contains direct and inclusive values for a measurement.
type Cost struct {
	Self       int64
	Cumulative int64
}

// CodeMetrics contains measurements attributed to source code.
type CodeMetrics map[Metric]Cost

// Global contains process-wide data that cannot be assigned to source code.
type Global struct {
	Process          runtimemetrics.RuntimeMetrics
	ProfileTotals    map[Metric]int64
	ProfileDurations map[string]time.Duration
	Trace            TraceSummary
}

// TraceSummary contains process-wide execution trace aggregates.
type TraceSummary struct {
	Duration        time.Duration
	Goroutines      uint64
	LiveGoroutines  uint64
	PeakGoroutines  uint64
	Processors      uint64
	Threads         uint64
	StackSamples    uint64
	GoroutineStates map[runtimetrace.State]time.Duration
	ProcessorStates map[runtimetrace.State]time.Duration
	Ranges          map[string]time.Duration
	Metrics         map[string]uint64
}

// RuntimeUpdate transfers ownership of one or more completed runtime snapshots.
type RuntimeUpdate struct {
	Reset    bool
	Metrics  *runtimemetrics.RuntimeMetrics
	Profiles map[string]*ProfileUpdate
	Trace    *TraceUpdate
}

// ProfileUpdate contains one profile's global and source-attributed costs.
type ProfileUpdate struct {
	Duration time.Duration
	Totals   map[Metric]int64
	Code     CodeUpdate
}

// TraceUpdate contains one trace's global and source-attributed aggregates.
type TraceUpdate struct {
	Summary TraceSummary
	Code    CodeUpdate
}

// CodeUpdate contains runtime costs grouped by static graph identity.
type CodeUpdate struct {
	Nodes    map[graph.NodeID]CodeMetrics
	Files    map[graph.FileID]CodeMetrics
	Unmapped map[Metric]int64
}

// BuildRuntimeGraph allocates the main module's stable model owned by the TUI.
func BuildRuntimeGraph(modulePath string, static *graph.Graph) *RuntimeGraph {
	result := &RuntimeGraph{
		Static:   static,
		Nodes:    make(map[graph.NodeID]*Node),
		Files:    make(map[graph.FileID]*File),
		Unmapped: make(map[Metric]int64),
		Global: Global{
			ProfileTotals:    make(map[Metric]int64),
			ProfileDurations: make(map[string]time.Duration),
			Trace:            newTraceSummary(),
		},
	}
	result.mapper = newMapper(modulePath, static, result)
	return result
}

// BuildUpdate aggregates a complete runtime snapshot without mutating the graph.
func (g *RuntimeGraph) BuildUpdate(observed *runtime.Runtime) RuntimeUpdate {
	update := RuntimeUpdate{Reset: true, Metrics: copyMetrics(observed.Metrics)}
	if len(observed.Profiles) != 0 {
		update.Profiles = acquireProfileMap()
		for name, profile := range observed.Profiles {
			update.Profiles[name] = g.buildProfile(name, profile)
		}
	}
	if observed.Trace != nil {
		update.Trace = g.buildTrace(observed.Trace)
	}
	return update
}

// BuildMetricsUpdate aggregates a process-metrics snapshot.
func (g *RuntimeGraph) BuildMetricsUpdate(metrics *runtimemetrics.RuntimeMetrics) RuntimeUpdate {
	return RuntimeUpdate{Metrics: copyMetrics(metrics)}
}

// BuildProfileUpdate aggregates one named profile.
func (g *RuntimeGraph) BuildProfileUpdate(name string, profile *runtimepprof.Profile) RuntimeUpdate {
	profiles := acquireProfileMap()
	profiles[name] = g.buildProfile(name, profile)
	return RuntimeUpdate{Profiles: profiles}
}

// BuildTraceUpdate aggregates one runtime trace.
func (g *RuntimeGraph) BuildTraceUpdate(trace *runtimetrace.Trace) RuntimeUpdate {
	return RuntimeUpdate{Trace: g.buildTrace(trace)}
}

// ApplyUpdate consumes an update and applies it to the TUI-owned graph.
func (g *RuntimeGraph) ApplyUpdate(update RuntimeUpdate) {
	if update.Reset {
		g.resetRuntime()
	}
	if update.Metrics != nil {
		assignMetrics(&g.Global.Process, update.Metrics)
	} else if update.Reset {
		g.Global.Process = runtimemetrics.RuntimeMetrics{}
	}
	for name, profile := range update.Profiles {
		if !update.Reset {
			g.clearSource(name)
			delete(g.Global.ProfileDurations, name)
		}
		if profile == nil {
			continue
		}
		g.Global.ProfileDurations[name] = profile.Duration
		maps.Copy(g.Global.ProfileTotals, profile.Totals)
		g.applyCode(profile.Code, update.Reset)
	}
	if update.Trace != nil {
		if !update.Reset {
			g.clearSource(TraceSource)
		}
		g.copyTrace(update.Trace.Summary)
		g.applyCode(update.Trace.Code, update.Reset)
	}
	releaseUpdate(update)
}

func (g *RuntimeGraph) resetRuntime() {
	for _, node := range g.Nodes {
		clear(node.Metrics)
	}
	for _, file := range g.Files {
		clear(file.Metrics)
	}
	clear(g.Unmapped)
	clear(g.Global.ProfileTotals)
	clear(g.Global.ProfileDurations)
	g.copyTrace(newTraceSummary())
}

func (g *RuntimeGraph) clearSource(source string) {
	for _, node := range g.Nodes {
		clearMetricsSource(node.Metrics, source)
	}
	for _, file := range g.Files {
		clearMetricsSource(file.Metrics, source)
	}
	clearMetricsSource(g.Unmapped, source)
	clearMetricsSource(g.Global.ProfileTotals, source)
}

func clearMetricsSource[T any](metrics map[Metric]T, source string) {
	for metric := range metrics {
		if metric.Source == source {
			delete(metrics, metric)
		}
	}
}

// applyCode transfers the first full-snapshot map and copies later sources.
func (g *RuntimeGraph) applyCode(update CodeUpdate, transfer bool) {
	for id, metrics := range update.Nodes {
		node := g.Nodes[id]
		if transfer && len(node.Metrics) == 0 {
			node.Metrics, update.Nodes[id] = metrics, node.Metrics
		} else {
			maps.Copy(node.Metrics, metrics)
		}
	}
	for id, metrics := range update.Files {
		file := g.Files[id]
		if transfer && len(file.Metrics) == 0 {
			file.Metrics, update.Files[id] = metrics, file.Metrics
		} else {
			maps.Copy(file.Metrics, metrics)
		}
	}
	maps.Copy(g.Unmapped, update.Unmapped)
}

func (g *RuntimeGraph) copyTrace(source TraceSummary) {
	target := &g.Global.Trace
	target.Duration = source.Duration
	target.Goroutines = source.Goroutines
	target.LiveGoroutines = source.LiveGoroutines
	target.PeakGoroutines = source.PeakGoroutines
	target.Processors = source.Processors
	target.Threads = source.Threads
	target.StackSamples = source.StackSamples
	clear(target.GoroutineStates)
	clear(target.ProcessorStates)
	clear(target.Ranges)
	clear(target.Metrics)
	maps.Copy(target.GoroutineStates, source.GoroutineStates)
	maps.Copy(target.ProcessorStates, source.ProcessorStates)
	maps.Copy(target.Ranges, source.Ranges)
	maps.Copy(target.Metrics, source.Metrics)
}

func newTraceSummary() TraceSummary {
	return TraceSummary{
		GoroutineStates: make(map[runtimetrace.State]time.Duration),
		ProcessorStates: make(map[runtimetrace.State]time.Duration),
		Ranges:          make(map[string]time.Duration),
		Metrics:         make(map[string]uint64),
	}
}

func newCodeUpdate() CodeUpdate {
	return CodeUpdate{
		Nodes:    make(map[graph.NodeID]CodeMetrics),
		Files:    make(map[graph.FileID]CodeMetrics),
		Unmapped: make(map[Metric]int64),
	}
}

func (u *CodeUpdate) add(metric Metric, value int64, nodes []graph.NodeID, files []graph.FileID) {
	if len(files) == 0 {
		u.Unmapped[metric] += value
	}
	for index, id := range nodes {
		metrics := u.Nodes[id]
		if metrics == nil {
			metrics = acquireCodeMetrics()
			u.Nodes[id] = metrics
		}
		cost := metrics[metric]
		if index == 0 {
			cost.Self += value
		}
		cost.Cumulative += value
		metrics[metric] = cost
	}
	for index, id := range files {
		metrics := u.Files[id]
		if metrics == nil {
			metrics = acquireCodeMetrics()
			u.Files[id] = metrics
		}
		cost := metrics[metric]
		if index == 0 {
			cost.Self += value
		}
		cost.Cumulative += value
		metrics[metric] = cost
	}
}

func (g *RuntimeGraph) buildProfile(name string, profile *runtimepprof.Profile) *ProfileUpdate {
	if profile == nil {
		return nil
	}
	update := acquireProfileUpdate()
	update.Duration = profile.Duration
	workspace := acquireTargetWorkspace()
	defer releaseTargetWorkspace(workspace)
	for _, sample := range profile.Samples {
		nodes, files := g.mapper.profileTargets(profile, sample, workspace)
		for index, value := range sample.Values {
			if index >= len(profile.SampleTypes) {
				break
			}
			sampleType := profile.SampleTypes[index]
			metric := Metric{Source: name, Name: sampleType.Type, Unit: sampleType.Unit}
			update.Totals[metric] += value
			update.Code.add(metric, value, nodes, files)
		}
	}
	return update
}
