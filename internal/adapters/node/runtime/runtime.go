// Package runtime manages a Node.js target and collects its runtime data.
package runtime

import (
	"context"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/node/runtime/cdp"
	"github.com/briheet/sen/internal/adapters/node/runtime/cpuprofile"
	"github.com/briheet/sen/internal/adapters/node/runtime/process"
	"github.com/briheet/sen/internal/adapters/processstats"
	"github.com/briheet/sen/internal/model"
)

const (
	cpuProfile    = "cpu"
	profileWindow = time.Second
)

// Runtime owns a Node.js target process and its collected data.
type Runtime struct {
	Process  *process.Process
	Metrics  *model.RuntimeMetrics
	Profiles map[string]*model.Profile
	Trace    *model.Trace

	client  *cdp.Client
	sampler *processstats.Sampler
}

var _ adapters.Runtime = (*Runtime)(nil)

// NewRuntime builds the target process for the given source directory.
func NewRuntime(ctx context.Context, sourcePath string, buildArgs, runArgs []string, output adapters.Output) (*Runtime, error) {
	target, err := process.NewProcess(ctx, sourcePath, buildArgs, runArgs, output)
	if err != nil {
		return nil, err
	}
	return &Runtime{
		Process:  target,
		Metrics:  &model.RuntimeMetrics{},
		Profiles: make(map[string]*model.Profile),
	}, nil
}

// Start launches the target and enables V8 profiling.
func (r *Runtime) Start(ctx context.Context) error {
	if err := r.Process.Start(); err != nil {
		return err
	}
	wsURL, err := r.Process.WaitURL(ctx)
	if err != nil {
		_ = r.Process.Stop()
		return err
	}
	client, err := cdp.Dial(ctx, wsURL)
	if err != nil {
		_ = r.Process.Stop()
		return err
	}
	if err := r.startProfiler(ctx, client); err != nil {
		_ = client.Close()
		_ = r.Process.Stop()
		return err
	}
	r.client = client
	sampler, err := processstats.New(ctx, r.Process.PID())
	if err != nil {
		_ = client.Close()
		_ = r.Process.Stop()
		return err
	}
	r.sampler = sampler
	return nil
}

// startProfiler begins sampling the target's V8 CPU profiler.
func (r *Runtime) startProfiler(ctx context.Context, client *cdp.Client) error {
	if err := client.Call(ctx, "Profiler.enable", nil, nil); err != nil {
		return err
	}
	if err := client.Call(ctx, "Profiler.setSamplingInterval", map[string]any{"interval": 1000}, nil); err != nil {
		return err
	}
	return client.Call(ctx, "Profiler.start", nil, nil)
}

// Collect captures one V8 profile window and current process metrics.
func (r *Runtime) Collect(ctx context.Context) (model.Observation, error) {
	timer := time.NewTimer(profileWindow)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return model.Observation{}, ctx.Err()
	case <-timer.C:
	}

	var response struct {
		Profile cpuprofile.CPUProfile `json:"profile"`
	}
	if err := r.client.Call(ctx, "Profiler.stop", nil, &response); err != nil {
		return model.Observation{}, err
	}
	if err := r.client.Call(ctx, "Profiler.start", nil, nil); err != nil {
		return model.Observation{}, err
	}
	r.Profiles[cpuProfile] = response.Profile.Profile()
	r.Trace = response.Profile.Trace()
	if err := r.collectMetrics(ctx); err != nil {
		return model.Observation{}, err
	}
	if r.sampler != nil {
		r.Metrics.Process = r.sampler.Collect(ctx)
	}
	return model.Observation{Metrics: r.Metrics, Profiles: r.Profiles, Trace: r.Trace}, nil
}

// Wait blocks until the target exits.
func (r *Runtime) Wait() error {
	err := r.Process.Wait()
	if r.client != nil {
		_ = r.client.Close()
	}
	return err
}

// Stop terminates the target.
func (r *Runtime) Stop() error {
	if r.client != nil {
		_ = r.client.Close()
	}
	return r.Process.Stop()
}

// Cleanup removes temporary target files.
func (r *Runtime) Cleanup() error {
	return r.Process.Cleanup()
}

// shimRow mirrors one runtime snapshot returned by the injected shim.
type shimRow struct {
	HeapUsed             uint64  `json:"heapUsed"`
	HeapTotal            uint64  `json:"heapTotal"`
	External             uint64  `json:"external"`
	ArrayBuffers         uint64  `json:"arrayBuffers"`
	EventLoopUtilization float64 `json:"eventLoopUtilization"`
	EventLoopDelayMean   float64 `json:"eventLoopDelayMean"`
	EventLoopDelayMax    float64 `json:"eventLoopDelayMax"`
	EventLoopDelayP95    float64 `json:"eventLoopDelayP95"`
	EventLoopDelayP99    float64 `json:"eventLoopDelayP99"`
	ActiveResources      uint64  `json:"activeResources"`
}

// collectMetrics asks the target for one runtime snapshot.
func (r *Runtime) collectMetrics(ctx context.Context) error {
	var response struct {
		Result struct {
			Value shimRow `json:"value"`
		} `json:"result"`
	}
	if err := r.client.Call(ctx, "Runtime.evaluate", map[string]any{
		"expression":    "globalThis[Symbol.for('sen.metrics')]()",
		"returnByValue": true,
	}, &response); err != nil {
		return err
	}
	row := response.Result.Value
	r.Metrics.Node = model.NodeMetrics{
		HeapUsed:             row.HeapUsed,
		HeapTotal:            row.HeapTotal,
		External:             row.External,
		ArrayBuffers:         row.ArrayBuffers,
		EventLoopUtilization: row.EventLoopUtilization,
		EventLoopDelayMean:   time.Duration(row.EventLoopDelayMean),
		EventLoopDelayMax:    time.Duration(row.EventLoopDelayMax),
		EventLoopDelayP95:    time.Duration(row.EventLoopDelayP95),
		EventLoopDelayP99:    time.Duration(row.EventLoopDelayP99),
		ActiveResources:      row.ActiveResources,
	}
	return nil
}
