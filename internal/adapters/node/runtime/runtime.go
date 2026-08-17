// Package runtime manages a Node.js target and collects its runtime data.
package runtime

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/node/runtime/cdp"
	"github.com/briheet/senbon/internal/adapters/node/runtime/cpuprofile"
	"github.com/briheet/senbon/internal/adapters/node/runtime/process"
	"github.com/briheet/senbon/internal/model"
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

	client *cdp.Client
	offset int64
}

var _ adapters.Runtime = (*Runtime)(nil)

// NewRuntime builds the target process for the given source directory.
func NewRuntime(ctx context.Context, sourcePath string) (*Runtime, error) {
	target, err := process.NewProcess(ctx, sourcePath)
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
	if metrics, ok := readMetrics(r.Process.MetricsFile, &r.offset); ok {
		r.Metrics = metrics
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

// shimRow mirrors one line of the metrics shim output.
type shimRow struct {
	HeapUsed  uint64 `json:"heapUsed"`
	HeapTotal uint64 `json:"heapTotal"`
	RSS       uint64 `json:"rss"`
	User      uint64 `json:"user"`
}

// readMetrics reads appended metrics lines and returns the latest complete row.
func readMetrics(path string, offset *int64) (*model.RuntimeMetrics, bool) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= *offset {
		return nil, false
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Seek(*offset, io.SeekStart); err != nil {
		return nil, false
	}
	data := make([]byte, info.Size()-*offset)
	n, err := io.ReadFull(file, data)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, false
	}
	*offset += int64(n)

	var latest model.RuntimeMetrics
	found := false
	for _, line := range strings.Split(string(data[:n]), "\n") {
		var row shimRow
		if json.Unmarshal([]byte(line), &row) != nil {
			continue
		}
		latest = model.RuntimeMetrics{
			LiveHeap:        row.HeapUsed,
			HeapAlloc:       row.HeapTotal,
			TotalRuntimeMem: row.RSS,
			UserCPU:         float64(row.User) / 1e6,
		}
		found = true
	}
	if !found {
		return nil, false
	}
	return &latest, true
}
