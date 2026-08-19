package engine

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/config"
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
