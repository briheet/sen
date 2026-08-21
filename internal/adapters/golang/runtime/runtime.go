// Package runtime manages the target process and its runtime data.
package runtime

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/briheet/sen/internal/adapters"
	runtimemetrics "github.com/briheet/sen/internal/adapters/golang/runtime/metrics"
	runtimepprof "github.com/briheet/sen/internal/adapters/golang/runtime/pprof"
	runtimeprocess "github.com/briheet/sen/internal/adapters/golang/runtime/process"
	runtimetrace "github.com/briheet/sen/internal/adapters/golang/runtime/trace"
	"github.com/briheet/sen/internal/adapters/processstats"
	"github.com/briheet/sen/internal/model"
)

// Runtime owns a target process and its collected runtime data.
type Runtime struct {
	Process  *runtimeprocess.Process
	Metrics  *runtimemetrics.RuntimeMetrics
	Profiles map[string]*runtimepprof.Profile
	Trace    *runtimetrace.Trace

	client       *http.Client
	sampler      *processstats.Sampler
	traceDecoder runtimetrace.Decoder
}

var _ adapters.Runtime = (*Runtime)(nil)

const (
	metricsPath   = "/debug/sen/metrics"
	pprofPath     = "/debug/pprof/"
	tracePath     = "/debug/pprof/trace"
	collectorURL  = "http://sen"
	cpuProfile    = "cpu"
	collectRetry  = 10 * time.Millisecond
	collectWindow = time.Second
)

// NewRuntime builds the target process.
func NewRuntime(ctx context.Context, sourceDir string, buildArgs, runArgs []string, output adapters.Output) (*Runtime, error) {
	process, err := runtimeprocess.NewProcess(ctx, sourceDir, buildArgs, runArgs, output)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", process.CollectorSocket)
	}}
	return &Runtime{
		Process:  process,
		Metrics:  &runtimemetrics.RuntimeMetrics{},
		Profiles: make(map[string]*runtimepprof.Profile),
		client:   &http.Client{Transport: transport},
	}, nil
}

// Start launches the instrumented target.
func (r *Runtime) Start(ctx context.Context) error {
	if err := r.Process.Start(); err != nil {
		return err
	}
	sampler, err := processstats.New(ctx, r.Process.PID())
	if err != nil {
		_ = r.Process.Stop()
		return err
	}
	r.sampler = sampler
	return nil
}

// Collect captures one complete runtime snapshot.
func (r *Runtime) Collect(ctx context.Context) (model.Observation, error) {
	retry := time.NewTicker(collectRetry)
	defer retry.Stop()
	for {
		if err := r.CollectMetrics(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return model.Observation{}, ctx.Err()
		case <-retry.C:
		}
	}

	if err := r.CollectTrace(ctx, collectWindow); err != nil {
		return model.Observation{}, err
	}
	if err := r.CollectMetrics(ctx); err != nil {
		return model.Observation{}, err
	}
	if r.sampler != nil {
		r.Metrics.Process = r.sampler.Collect(ctx)
	}
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
	r.client.CloseIdleConnections()
	return r.Process.Cleanup()
}

// ReadMetrics decodes and stores target runtime metrics.
func (r *Runtime) ReadMetrics(reader io.Reader) error {
	result, err := runtimemetrics.Read(reader)
	if err != nil {
		return err
	}
	r.Metrics = result
	return nil
}

// ReadProfile decodes and stores a named pprof profile.
func (r *Runtime) ReadProfile(name string, reader io.Reader) error {
	profile, err := runtimepprof.Read(reader)
	if err != nil {
		return err
	}
	r.Profiles[name] = profile
	return nil
}

// ReadTrace decodes and stores trace data from the target process.
func (r *Runtime) ReadTrace(reader io.Reader) error {
	result, err := r.traceDecoder.Read(reader)
	if err != nil {
		return err
	}
	r.Trace = result
	return nil
}

// CollectMetrics requests current metrics from the target process.
func (r *Runtime) CollectMetrics(ctx context.Context) error {
	return r.collect(ctx, metricsPath, r.ReadMetrics)
}

// CollectProfile requests and stores a named pprof profile.
func (r *Runtime) CollectProfile(ctx context.Context, name string, duration time.Duration) error {
	path := pprofPath + url.PathEscape(name)
	if name == cpuProfile {
		path = pprofPath + "profile"
	}
	if duration > 0 {
		seconds := (duration + time.Second - 1) / time.Second
		path += "?seconds=" + strconv.FormatInt(int64(seconds), 10)
	}
	return r.collect(ctx, path, func(reader io.Reader) error {
		return r.ReadProfile(name, reader)
	})
}

// CollectTrace requests and stores a completed runtime trace.
func (r *Runtime) CollectTrace(ctx context.Context, duration time.Duration) error {
	path := tracePath
	if duration > 0 {
		seconds := (duration + time.Second - 1) / time.Second
		path += "?seconds=" + strconv.FormatInt(int64(seconds), 10)
	}
	return r.collect(ctx, path, r.ReadTrace)
}

func (r *Runtime) collect(ctx context.Context, path string, decode func(io.Reader) error) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, collectorURL+path, nil)
	if err != nil {
		return err
	}
	response, err := r.client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("collector returned %s", response.Status)
	}
	return decode(response.Body)
}
