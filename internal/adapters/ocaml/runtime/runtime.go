// Package runtime manages an OCaml target and collects its runtime data.
package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/ocaml/runtime/process"
	"github.com/briheet/senbon/internal/model"
)

const (
	cpuProfile    = "cpu"
	profileWindow = time.Second
)

// Runtime owns an OCaml target process and its collected data.
type Runtime struct {
	Process  *process.Process
	Metrics  *model.RuntimeMetrics
	Profiles map[string]*model.Profile
	Trace    *model.Trace
}

var _ adapters.Runtime = (*Runtime)(nil)

// NewRuntime builds the target process for the given entry source.
func NewRuntime(ctx context.Context, sourceDir, entry string) (*Runtime, error) {
	target, err := process.NewProcess(ctx, sourceDir, entry)
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

// Collect waits for the target to exit (flushing the events ring), then reads
// the runtime events buffer. OCaml only flushes the .events ring on process
// exit, so collection reflects the full process run.
func (r *Runtime) Collect(ctx context.Context) (model.Observation, error) {
	exited := make(chan error, 1)
	go func() { exited <- r.Process.Wait() }()
	select {
	case err := <-exited:
		if err != nil {
			return model.Observation{}, err
		}
	case <-ctx.Done():
		return model.Observation{}, ctx.Err()
	}

	file := r.findEvents()
	if file == "" {
		return model.Observation{}, errors.New("ocaml: no runtime events file produced")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return model.Observation{}, err
	}
	counts, err := decodeEvents(data)
	if err != nil {
		return model.Observation{}, err
	}

	r.Metrics.GCCycles = counts.MinorCollections + counts.MajorCollections
	names := loadFunctionMap(r.Process.FunctionMap)

	// Per-function call-count profile.
	profile := &model.Profile{
		SampleTypes: []model.ValueType{{Type: "calls", Unit: "count"}},
		PeriodType:  model.ValueType{Type: "calls", Unit: "count"},
		Period:      1,
		Locations:   make(map[model.ProfileLocationID]model.ProfileLocation),
	}
	for id, count := range counts.FunctionSpans {
		locID := model.ProfileLocationID(len(profile.Locations) + 1)
		name := names[id]
		if name == "" {
			name = itoa(int(id))
		}
		profile.Locations[locID] = model.ProfileLocation{
			ID:     locID,
			Frames: []model.ProfileFrame{{Function: name}},
		}
		if count > 0 {
			profile.Samples = append(profile.Samples, model.ProfileSample{
				Values: []int64{int64(count)},
				Stack:  []model.ProfileLocationID{locID},
			})
		}
	}
	profile.Duration = profileWindow
	r.Profiles["ocaml"] = profile

	// Per-function execution timeline (spans, in emission order).
	trace := &model.Trace{Stacks: make(map[model.StackID]model.TraceStack)}
	trace.Duration = profileWindow
	stackByName := make(map[string]model.StackID, len(counts.FunctionSpans))
	for _, span := range counts.Spans {
		name := names[span.Function]
		if name == "" {
			name = itoa(int(span.Function))
		}
		stackID, ok := stackByName[name]
		if !ok {
			stackID = model.StackID(len(trace.Stacks) + 1)
			trace.Stacks[stackID] = model.TraceStack{Frames: []model.TraceFrame{{Function: name}}}
			stackByName[name] = stackID
		}
		trace.Events = append(trace.Events, model.Event{
			At:    time.Duration(span.Timestamp),
			Kind:  model.EventStackSample,
			Stack: stackID,
		})
	}
	r.Trace = trace

	return model.Observation{Metrics: r.Metrics, Profiles: r.Profiles, Trace: r.Trace}, nil
}

// loadFunctionMap reads the instrument id->name JSON emitted by instrument.ml.
func loadFunctionMap(path string) map[uint64]string {
	result := make(map[uint64]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return result
	}
	for key, name := range raw {
		if id, err := strconv.ParseUint(key, 10, 64); err == nil {
			result[id] = name
		}
	}
	return result
}

// findEvents locates the target's runtime events ring buffer.
func (r *Runtime) findEvents() string {
	if entries, err := os.ReadDir(r.Process.EventsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && len(entry.Name()) > len(".events") &&
				entry.Name()[len(entry.Name())-len(".events"):] == ".events" {
				return r.Process.EventsDir + "/" + entry.Name()
			}
		}
	}
	return ""
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

// itoa formats a function id for fallback naming.
func itoa(n int) string {
	return strconv.Itoa(n)
}
