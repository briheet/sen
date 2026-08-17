package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEngineBuildsRuntimeGraph(t *testing.T) {
	sourceDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example.com/app\n\ngo 1.25\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600))

	engine, err := NewEngine(context.Background(), sourceDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Runtime.Process.Cleanup() })
	require.NotNil(t, engine.Runtime)
	require.NotNil(t, engine.Graph)
	require.NotNil(t, engine.Graph.Static)
	require.NotEmpty(t, engine.Graph.Nodes)
	require.NotEmpty(t, engine.Graph.Files)

	graph := engine.Graph
	engine.Runtime.Metrics.UserCPU = 1.5
	update := engine.MetricsUpdate()
	require.Same(t, graph, engine.Graph)
	require.Zero(t, engine.Graph.Global.Process.UserCPU)
	engine.Graph.ApplyUpdate(update)
	require.Equal(t, 1.5, engine.Graph.Global.Process.UserCPU)

	engine.Runtime.Metrics.UserCPU = 2.5
	update = engine.Snapshot()
	require.NotNil(t, update.Metrics)
	engine.Graph.ApplyUpdate(update)
	require.Equal(t, 2.5, engine.Graph.Global.Process.UserCPU)
}
