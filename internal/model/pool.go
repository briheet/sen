package model

import (
	"sync"

	runtimemetrics "github.com/briheet/senbon/internal/runtime/metrics"
	stdruntimemetrics "runtime/metrics"
)

const maxPooledMapEntries = 4096

// Update pools retain map buckets until ApplyUpdate consumes their payload.
var (
	profileUpdates = sync.Pool{New: func() any {
		return &ProfileUpdate{Totals: make(map[Metric]int64), Code: newCodeUpdate()}
	}}
	traceUpdates = sync.Pool{New: func() any {
		return &TraceUpdate{Summary: newTraceSummary(), Code: newCodeUpdate()}
	}}
	profileMaps = sync.Pool{New: func() any {
		return make(map[string]*ProfileUpdate)
	}}
	codeMetricsMaps = sync.Pool{New: func() any {
		return make(CodeMetrics)
	}}
	metricsUpdates = sync.Pool{New: func() any {
		return new(runtimemetrics.RuntimeMetrics)
	}}
	histogramUpdates = sync.Pool{New: func() any {
		return new(stdruntimemetrics.Float64Histogram)
	}}
)

func acquireProfileUpdate() *ProfileUpdate {
	return profileUpdates.Get().(*ProfileUpdate)
}

func acquireTraceUpdate() *TraceUpdate {
	return traceUpdates.Get().(*TraceUpdate)
}

func acquireProfileMap() map[string]*ProfileUpdate {
	return profileMaps.Get().(map[string]*ProfileUpdate)
}

func acquireCodeMetrics() CodeMetrics {
	return codeMetricsMaps.Get().(CodeMetrics)
}

func copyMetrics(source *runtimemetrics.RuntimeMetrics) *runtimemetrics.RuntimeMetrics {
	target := metricsUpdates.Get().(*runtimemetrics.RuntimeMetrics)
	*target = *source
	target.SchedulerLatency = copyHistogram(source.SchedulerLatency)
	target.GCPauses = copyHistogram(source.GCPauses)
	return target
}

func copyHistogram(source *stdruntimemetrics.Float64Histogram) *stdruntimemetrics.Float64Histogram {
	if source == nil {
		return nil
	}
	target := histogramUpdates.Get().(*stdruntimemetrics.Float64Histogram)
	target.Counts = append(target.Counts[:0], source.Counts...)
	target.Buckets = append(target.Buckets[:0], source.Buckets...)
	return target
}

func assignMetrics(target, source *runtimemetrics.RuntimeMetrics) {
	schedulerLatency := target.SchedulerLatency
	gcPauses := target.GCPauses
	*target = *source
	target.SchedulerLatency = assignHistogram(schedulerLatency, source.SchedulerLatency)
	target.GCPauses = assignHistogram(gcPauses, source.GCPauses)
}

func assignHistogram(target, source *stdruntimemetrics.Float64Histogram) *stdruntimemetrics.Float64Histogram {
	if source == nil {
		return nil
	}
	if target == nil {
		target = new(stdruntimemetrics.Float64Histogram)
	}
	target.Counts = append(target.Counts[:0], source.Counts...)
	target.Buckets = append(target.Buckets[:0], source.Buckets...)
	return target
}

func releaseUpdate(update RuntimeUpdate) {
	if update.Metrics != nil {
		releaseMetrics(update.Metrics)
	}
	for _, profile := range update.Profiles {
		if profile != nil {
			releaseProfileUpdate(profile)
		}
	}
	if update.Profiles != nil {
		resetMap(&update.Profiles)
		profileMaps.Put(update.Profiles)
	}
	if update.Trace != nil {
		releaseTraceUpdate(update.Trace)
	}
}

func releaseProfileUpdate(update *ProfileUpdate) {
	update.Duration = 0
	resetMap(&update.Totals)
	releaseCodeUpdate(&update.Code)
	profileUpdates.Put(update)
}

func releaseTraceUpdate(update *TraceUpdate) {
	resetTraceSummary(&update.Summary)
	releaseCodeUpdate(&update.Code)
	traceUpdates.Put(update)
}

func releaseCodeUpdate(update *CodeUpdate) {
	for _, metrics := range update.Nodes {
		releaseCodeMetrics(metrics)
	}
	for _, metrics := range update.Files {
		releaseCodeMetrics(metrics)
	}
	resetMap(&update.Nodes)
	resetMap(&update.Files)
	resetMap(&update.Unmapped)
}

func releaseCodeMetrics(metrics CodeMetrics) {
	if len(metrics) > maxPooledMapEntries {
		return
	}
	clear(metrics)
	codeMetricsMaps.Put(metrics)
}

func releaseMetrics(metrics *runtimemetrics.RuntimeMetrics) {
	releaseHistogram(metrics.SchedulerLatency)
	releaseHistogram(metrics.GCPauses)
	*metrics = runtimemetrics.RuntimeMetrics{}
	metricsUpdates.Put(metrics)
}

func releaseHistogram(histogram *stdruntimemetrics.Float64Histogram) {
	if histogram == nil {
		return
	}
	if cap(histogram.Counts) > maxPooledMapEntries || cap(histogram.Buckets) > maxPooledMapEntries {
		histogram.Counts = nil
		histogram.Buckets = nil
	}
	histogramUpdates.Put(histogram)
}

func resetTraceSummary(summary *TraceSummary) {
	summary.Duration = 0
	summary.Goroutines = 0
	summary.LiveGoroutines = 0
	summary.PeakGoroutines = 0
	summary.Processors = 0
	summary.Threads = 0
	summary.StackSamples = 0
	resetMap(&summary.GoroutineStates)
	resetMap(&summary.ProcessorStates)
	resetMap(&summary.Ranges)
	resetMap(&summary.Metrics)
}

func resetMap[K comparable, V any](target *map[K]V) {
	if len(*target) > maxPooledMapEntries {
		*target = make(map[K]V)
		return
	}
	clear(*target)
}
