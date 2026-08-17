package runtime_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/briheet/senbon/internal/adapters/ocaml/runtime"
	"github.com/stretchr/testify/require"
)

const fixtureApp = `let rec fib n = if n < 2 then n else fib (n - 1) + fib (n - 2)
let () =
  let a = Array.init 100000 (fun i -> i mod 1000) in
  let _ = Array.fold_left (fun s x -> s + fib (x mod 20)) 0 a in
  Gc.compact ();
  print_endline "done"
`

// TestCollect verifies OCaml GC/alloc runtime data is collected.
func TestCollect(t *testing.T) {
	if _, err := exec.LookPath("ocamlc"); err != nil {
		t.Skip("ocamlc not installed")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "main.ml")
	require.NoError(t, os.WriteFile(fixture, []byte(fixtureApp), 0o600))

	rt, err := runtime.NewRuntime(context.Background(), dir, fixture)
	require.NoError(t, err)
	defer func() { _ = rt.Cleanup() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	require.NoError(t, rt.Start(ctx))

	observation, err := rt.Collect(ctx)
	require.NoError(t, err)
	require.NotZero(t, observation.Metrics.GCCycles, "expected GC cycles from runtime events")
	require.NotNil(t, observation.Trace)
	fmt.Fprintln(os.Stderr, "gc cycles:", observation.Metrics.GCCycles)
	require.NoError(t, rt.Stop())
}
