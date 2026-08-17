package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/stretchr/testify/require"
)

func TestNewEngineBuildsRuntimeGraph(t *testing.T) {
	sourceDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example.com/app\n\ngo 1.25\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600))

	engine, err := NewEngine(context.Background(), sourceDir, adapters.GoTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Cleanup() })
	require.NotNil(t, engine.Runtime)
	require.NotNil(t, engine.Graph)
	require.NotNil(t, engine.Graph.Static)
	require.NotEmpty(t, engine.Graph.Nodes)
	require.NotEmpty(t, engine.Graph.Files)
}
