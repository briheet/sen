// Package runtime collects native Rust process and CPU telemetry.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/processstats"
	"github.com/briheet/sen/internal/adapters/rust/analysis"
	rustbuild "github.com/briheet/sen/internal/adapters/rust/build"
	"github.com/briheet/sen/internal/adapters/rust/runtime/console"
	"github.com/briheet/sen/internal/adapters/rust/runtime/process"
	"github.com/briheet/sen/internal/adapters/rust/runtime/profiler"
	"github.com/briheet/sen/internal/model"
)

const cpuProfile = "cpu"

// Runtime owns one prepared Rust process and strict native profiler.
type Runtime struct {
	process *process.Process
	symbols *analysis.Symbols
	sampler *processstats.Sampler
	pending *model.Observation
	console *console.Collector
}

var _ adapters.Runtime = (*Runtime)(nil)

// New creates a Rust runtime without starting its child process.
func New(prepared *rustbuild.Prepared, symbols *analysis.Symbols, runArgs []string, output adapters.Output) *Runtime {
	return &Runtime{process: process.New(prepared, runArgs, output), symbols: symbols}
}

// Start launches and immediately validates the native profiler with a full window.
func (r *Runtime) Start(ctx context.Context) error {
	var consoleAddress string
	if r.process.Prepared.ConsoleMode != "off" {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("reserve Tokio Console address: %w", err)
		}
		consoleAddress = listener.Addr().String()
		_ = listener.Close()
		r.process.AddEnv(
			"TOKIO_CONSOLE_BIND="+consoleAddress,
			"TOKIO_CONSOLE_PUBLISH_INTERVAL=1s",
		)
	}
	if err := r.process.Start(); err != nil {
		return err
	}
	if consoleAddress != "" {
		collector, err := console.Connect(ctx, consoleAddress)
		if err != nil {
			_ = r.process.Stop()
			return err
		}
		r.console = collector
	}
	sampler, err := processstats.New(ctx, r.process.PID())
	if err != nil {
		_ = r.process.Stop()
		return err
	}
	r.sampler = sampler
	observation, err := r.capture(ctx)
	if err != nil {
		_ = r.process.Stop()
		return err
	}
	r.pending = &observation
	return nil
}

// Collect returns the preflight window once, then captures rolling one-second windows.
func (r *Runtime) Collect(ctx context.Context) (model.Observation, error) {
	if r.pending != nil {
		result := *r.pending
		r.pending = nil
		return result, nil
	}
	return r.capture(ctx)
}

func (r *Runtime) capture(ctx context.Context) (model.Observation, error) {
	if r.console != nil {
		if err := r.console.Err(); err != nil {
			return model.Observation{}, err
		}
	}
	profile, trace, count, err := profiler.Capture(ctx, r.process.PID(), r.symbols)
	if err != nil {
		return model.Observation{}, err
	}
	metrics := &model.RuntimeMetrics{Rust: model.RustMetrics{ProfileSamples: count, ProfileStacks: uint64(len(trace.Stacks))}}
	if r.console != nil {
		tokio := r.console.Snapshot()
		tokio.ProfileSamples = count
		tokio.ProfileStacks = uint64(len(trace.Stacks))
		metrics.Rust = tokio
	}
	if r.sampler != nil {
		metrics.Process = r.sampler.Collect(ctx)
	}
	return model.Observation{Metrics: metrics, Profiles: map[string]*model.Profile{cpuProfile: profile}, Trace: trace}, nil
}

func (r *Runtime) Wait() error { return r.process.Wait() }
func (r *Runtime) Stop() error {
	var consoleErr error
	if r.console != nil {
		consoleErr = r.console.Close()
	}
	return errors.Join(consoleErr, r.process.Stop())
}
func (r *Runtime) Cleanup() error { return r.process.Cleanup() }
