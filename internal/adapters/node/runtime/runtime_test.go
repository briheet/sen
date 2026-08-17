package runtime

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

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

	runtime, err := NewRuntime(context.Background(), dir)
	require.NoError(t, err)
	defer func() { _ = runtime.Cleanup() }()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	require.NoError(t, runtime.Start(ctx))
	done := make(chan error, 1)
	go func() { done <- runtime.Wait() }()

	observation, err := runtime.Collect(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, observation.Profiles["cpu"].Samples)
	require.Positive(t, observation.Metrics.LiveHeap)
	require.NotEmpty(t, observation.Trace.Events)

	require.NoError(t, runtime.Stop())
	<-done
}
