package build

import (
	"context"
	"errors"
	"testing"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRequiresRunnableService(t *testing.T) {
	configuration := &config.Config{
		Project:  config.Project{Name: "test"},
		Services: []config.Service{{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis, Address: "localhost:6379"}},
	}
	_, err := New(context.Background(), configuration)
	require.ErrorIs(t, err, errNoRunnableServices)
}

func TestGroupRunWaitsForAll(t *testing.T) {
	first := make(chan error, 1)
	second := make(chan error, 1)
	group := &Group{
		Engines: []*engine.Engine{testEngine("api", first), testEngine("worker", second)},
		logger:  zap.NewNop(),
	}
	done := make(chan error, 1)
	go func() { done <- group.Run() }()

	first <- nil
	select {
	case <-done:
		t.Fatal("Group.Run returned before every engine exited")
	default:
	}
	want := errors.New("worker failed")
	second <- want
	require.ErrorIs(t, <-done, want)
}

func TestGroupStopStopsEveryEngine(t *testing.T) {
	first := make(chan struct{}, 1)
	second := make(chan struct{}, 1)
	group := &Group{
		Engines: []*engine.Engine{testEngineWithStop("api", first), testEngineWithStop("worker", second)},
		logger:  zap.NewNop(),
	}

	require.NoError(t, group.Stop())
	<-first
	<-second
}

func testEngine(name string, wait <-chan error) *engine.Engine {
	static := &model.StaticGraph{
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	return &engine.Engine{
		Service: config.Service{Name: name},
		Runtime: &runtimeStub{wait: wait},
		Graph:   model.BuildRuntimeGraph("", static),
	}
}

func testEngineWithStop(name string, stopped chan<- struct{}) *engine.Engine {
	target := testEngine(name, nil)
	target.Runtime.(*runtimeStub).stopped = stopped
	return target
}

type runtimeStub struct {
	wait    <-chan error
	stopped chan<- struct{}
}

var _ adapters.Runtime = (*runtimeStub)(nil)

func (*runtimeStub) Start(context.Context) error { return nil }

func (*runtimeStub) Collect(context.Context) (model.Observation, error) {
	return model.Observation{Metrics: new(model.RuntimeMetrics)}, nil
}

func (r *runtimeStub) Wait() error { return <-r.wait }

func (r *runtimeStub) Stop() error {
	if r.stopped != nil {
		r.stopped <- struct{}{}
	}
	return nil
}

func (*runtimeStub) Cleanup() error { return nil }
