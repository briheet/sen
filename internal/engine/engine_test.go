package engine

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewEngineBuildsRuntimeGraph(t *testing.T) {
	sourceDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example.com/app\n\ngo 1.25\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("//go:build configured\n\npackage main\n\nfunc main() {}\n"), 0o600))

	engine, err := NewEngine(context.Background(), config.Service{
		Name:      "api",
		Type:      config.ServiceTypeServer,
		Lang:      config.ServiceLangGo,
		Path:      sourceDir,
		BuildArgs: []string{"-tags=configured"},
	}, adapters.Output{Stdout: io.Discard, Stderr: io.Discard})
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Cleanup() })
	require.Equal(t, "api", engine.Service.Name)
	require.NotNil(t, engine.Runtime)
	require.NotNil(t, engine.Graph)
	require.NotNil(t, engine.Graph.Static)
	require.NotEmpty(t, engine.Graph.Nodes)
	require.NotEmpty(t, engine.Graph.Files)
}

func TestNewEngineBuildsRedisRuntimeGraph(t *testing.T) {
	target, err := NewEngine(context.Background(), config.Service{
		Name:     "cache",
		Type:     config.ServiceTypeKV,
		Provider: config.ServiceProviderRedis,
		Address:  "localhost:6379",
	}, adapters.Output{Stdout: io.Discard, Stderr: io.Discard})
	require.NoError(t, err)
	t.Cleanup(func() { _ = target.Cleanup() })
	require.Equal(t, "cache", target.Service.Name)
	require.NotNil(t, target.Runtime)
	require.NotNil(t, target.Graph)
	require.NotEmpty(t, target.Graph.Nodes)
}

func TestNewEngineBuildsTigerBeetleRuntimeGraph(t *testing.T) {
	target, err := NewEngine(context.Background(), config.Service{
		Name: "ledger", Type: config.ServiceTypeDB, Provider: config.ServiceProviderTigerBeetle,
		Addresses: []string{"127.0.0.1:3000", "127.0.0.1:3001"}, MetricsAddress: "127.0.0.1:8125",
	}, adapters.Output{Stdout: io.Discard, Stderr: io.Discard})
	require.NoError(t, err)
	t.Cleanup(func() { _ = target.Cleanup() })
	require.NotNil(t, target.Runtime)
	require.Len(t, target.Graph.Static.Nodes, 11)
}

func TestRunCollectsUntilProcessExit(t *testing.T) {
	runtime := &continuousRuntime{exit: make(chan error, 1)}
	graph := model.BuildRuntimeGraph("", &model.StaticGraph{
		Nodes: make(map[model.NodeID]*model.StaticNode), Files: make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	})
	target := &Engine{Runtime: runtime, Graph: graph}

	require.NoError(t, target.Run())
	require.GreaterOrEqual(t, runtime.collections.Load(), int64(3))
	require.GreaterOrEqual(t, target.Revision(), uint64(3))
}

type continuousRuntime struct {
	exit        chan error
	collections atomic.Int64
}

func (*continuousRuntime) Start(context.Context) error { return nil }

func (r *continuousRuntime) Collect(ctx context.Context) (model.Observation, error) {
	select {
	case <-ctx.Done():
		return model.Observation{}, ctx.Err()
	case <-time.After(5 * time.Millisecond):
	}
	count := r.collections.Add(1)
	if count == 3 {
		r.exit <- nil
	}
	return model.Observation{Metrics: &model.RuntimeMetrics{Go: model.GoMetrics{GCCycles: uint64(count)}}}, nil
}

func (r *continuousRuntime) Wait() error  { return <-r.exit }
func (*continuousRuntime) Stop() error    { return nil }
func (*continuousRuntime) Cleanup() error { return nil }
