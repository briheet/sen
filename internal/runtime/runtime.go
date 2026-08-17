// Package runtime manages the target process and its runtime data.
package runtime

import (
	"context"
	"io"

	runtimemetrics "github.com/briheet/senbon/internal/runtime/metrics"
	runtimepprof "github.com/briheet/senbon/internal/runtime/pprof"
	runtimeprocess "github.com/briheet/senbon/internal/runtime/process"
	runtimetrace "github.com/briheet/senbon/internal/runtime/trace"
)

// Runtime owns a target process and its collected runtime data.
type Runtime struct {
	Process  *runtimeprocess.Process
	Metrics  *runtimemetrics.RuntimeMetrics
	Profiles map[string]*runtimepprof.Profile
	Trace    *runtimetrace.Trace
}

// NewRuntime builds the target process.
func NewRuntime(ctx context.Context, sourceDir string) (*Runtime, error) {
	process, err := runtimeprocess.NewProcess(ctx, sourceDir)
	if err != nil {
		return nil, err
	}
	return &Runtime{
		Process:  process,
		Metrics:  &runtimemetrics.RuntimeMetrics{},
		Profiles: make(map[string]*runtimepprof.Profile),
	}, nil
}

// ReadProfile decodes and stores a named pprof profile.
func (r *Runtime) ReadProfile(name string, reader io.Reader) error {
	profile, err := runtimepprof.Read(reader)
	if err != nil {
		return err
	}
	if r.Profiles == nil {
		r.Profiles = make(map[string]*runtimepprof.Profile)
	}
	r.Profiles[name] = profile
	return nil
}

// ReadTrace decodes and stores trace data from the target process.
func (r *Runtime) ReadTrace(reader io.Reader) error {
	result, err := runtimetrace.Read(reader)
	if err != nil {
		return err
	}
	r.Trace = result
	return nil
}
