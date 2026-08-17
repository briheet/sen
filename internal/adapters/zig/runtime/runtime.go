// Package runtime manages a Zig target and collects its runtime data.
package runtime

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/zig/analysis"
	"github.com/briheet/senbon/internal/adapters/zig/runtime/process"
	"github.com/briheet/senbon/internal/model"
)

const (
	cpuProfile    = "cpu"
	profileWindow = time.Second
)

// Runtime owns a Zig target process and its collected data.
type Runtime struct {
	Process  *process.Process
	Metrics  *model.RuntimeMetrics
	Profiles map[string]*model.Profile
	Trace    *model.Trace
}

var _ adapters.Runtime = (*Runtime)(nil)

// NewRuntime builds the target process for the given project.
func NewRuntime(ctx context.Context, sourcePath string, project *analysis.Project) (*Runtime, error) {
	target, err := process.NewProcess(ctx, sourcePath, project)
	if err != nil {
		return nil, err
	}
	return &Runtime{
		Process:  target,
		Metrics:  &model.RuntimeMetrics{},
		Profiles: make(map[string]*model.Profile),
	}, nil
}

// Start launches the target.
func (r *Runtime) Start(context.Context) error {
	return r.Process.Start()
}

// Collect captures one profiling window from the sampler file.
func (r *Runtime) Collect(ctx context.Context) (model.Observation, error) {
	timer := time.NewTimer(profileWindow)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return model.Observation{}, ctx.Err()
	case <-timer.C:
	}

	data, err := os.ReadFile(r.Process.SamplesFile)
	if err != nil {
		return model.Observation{}, err
	}
	base, interval := parseHeader(data)
	if base == 0 || interval == 0 {
		return model.Observation{}, errors.New("no samples collected")
	}
	symbolize, err := newSymbolizer(r.Process.BinPath, r.Process.DsymPath, base)
	if err != nil {
		return model.Observation{}, err
	}
	profile, trace := decodeSamples(data, interval, symbolize.line)
	r.Profiles[cpuProfile] = profile
	r.Trace = trace
	return model.Observation{Metrics: r.Metrics, Profiles: r.Profiles, Trace: r.Trace}, nil
}

// Wait blocks until the target exits.
func (r *Runtime) Wait() error {
	return r.Process.Wait()
}

// Stop terminates the target.
func (r *Runtime) Stop() error {
	return r.Process.Stop()
}

// Cleanup removes temporary target files.
func (r *Runtime) Cleanup() error {
	return r.Process.Cleanup()
}

// parseHeader extracts the marker base and sampling interval.
func parseHeader(data []byte) (base uint64, interval time.Duration) {
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "#base "):
			base, _ = strconv.ParseUint(strings.TrimPrefix(strings.TrimSpace(line[len("#base "):]), "0x"), 16, 64)
		case strings.HasPrefix(line, "#interval "):
			if ns, err := strconv.ParseInt(strings.TrimSpace(line[len("#interval "):]), 10, 64); err == nil {
				interval = time.Duration(ns)
			}
		}
	}
	return base, interval
}
