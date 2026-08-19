package runtime

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/require"
)

const fixtureApp = `function burn() {
  let x = 0;
  for (let i = 0; i < 1000000; i++) x += i;
  return x;
}
setInterval(burn, 10);
`

// TestCollect verifies Node data is normalized into one observation.
func TestCollect(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not installed")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.js"), []byte(fixtureApp), 0o600))

	var stdout, stderr bytes.Buffer
	runtime, err := NewRuntime(context.Background(), dir, nil, nil, adapters.Output{Stdout: &stdout, Stderr: &stderr})
	require.NoError(t, err)
	defer func() { _ = runtime.Cleanup() }()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	done := make(chan error, 1)
	go func() { done <- runtime.Wait() }()

	observation, err := runtime.Collect(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, observation.Profiles["cpu"].Samples)
	require.Positive(t, observation.Metrics.Node.HeapUsed)
	require.Positive(t, observation.Metrics.Node.External)
	require.Positive(t, observation.Metrics.Node.EventLoopDelayP99)
	require.GreaterOrEqual(t, observation.Metrics.Node.EventLoopUtilization, 0.0)
	require.Positive(t, observation.Metrics.Node.ActiveResources)
	require.True(t, observation.Metrics.Process.Has(model.ProcessCPU))
	require.True(t, observation.Metrics.Process.Has(model.ProcessMemory))
	require.NotEmpty(t, observation.Trace.Events)

	require.NoError(t, runtime.Stop())
	<-done
}
