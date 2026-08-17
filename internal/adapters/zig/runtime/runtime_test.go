package runtime

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/briheet/senbon/internal/adapters/zig/analysis"
	"github.com/stretchr/testify/require"
)

const fixtureApp = `const std = @import("std");

fn burn(deadline: usize) void {
    var sum: u64 = 0;
    var i: usize = 0;
    while (i < deadline) : (i += 1) {
        var j: u32 = 0;
        while (j < 200000) : (j += 1) sum +%= j;
    }
    std.debug.print("sum={d}\n", .{sum});
}

pub fn main() void {
    burn(3000);
}
`

// TestCollect verifies Zig runtime data is normalized into one observation.
func TestCollect(t *testing.T) {
	if _, err := exec.LookPath("zig"); err != nil {
		t.Skip("zig not installed")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.zig"), []byte(fixtureApp), 0o600))

	project, err := analysis.Analyze(context.Background(), dir)
	require.NoError(t, err)
	require.NotEmpty(t, project.Graph.Nodes)

	runtime, err := NewRuntime(context.Background(), dir, project)
	require.NoError(t, err)
	defer func() { _ = runtime.Cleanup() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	require.NoError(t, runtime.Start(ctx))
	done := make(chan error, 1)
	go func() { done <- runtime.Wait() }()

	observation, err := runtime.Collect(ctx)
	require.NoError(t, err)
	profile := observation.Profiles["cpu"]
	require.NotNil(t, profile)
	require.NotEmpty(t, profile.Samples)
	require.NotEmpty(t, profile.Locations)
	require.NotEmpty(t, observation.Trace.Events)

	mapped := false
	for _, location := range profile.Locations {
		if len(location.Frames) > 0 && strings.Contains(location.Frames[0].File, "main.zig") {
			mapped = true
			break
		}
	}
	require.True(t, mapped, "samples should symbolize to the target source")

	require.NoError(t, runtime.Stop())
	<-done
}
