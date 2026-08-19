package process

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/briheet/sen/internal/adapters"
	"github.com/stretchr/testify/require"
)

func TestNewProcessArguments(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "index.js")
	require.NoError(t, os.WriteFile(entry, []byte(""), 0o600))

	target, err := NewProcess(
		context.Background(),
		dir,
		[]string{"--no-warnings"},
		[]string{"--port", "8080"},
		adapters.Output{Stdout: io.Discard, Stderr: io.Discard},
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = target.Cleanup() })
	resolvedDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	require.Equal(t, []string{
		"node",
		"--inspect=" + inspectAddr,
		"--require", filepath.Join(target.tempDir, "shim.cjs"),
		"--no-warnings",
		filepath.Join(resolvedDir, filepath.Base(entry)),
		"--port", "8080",
	}, target.cmd.Args)
}

func TestScanStderrForwardsOutputAndFindsURL(t *testing.T) {
	var stderr bytes.Buffer
	target := &Process{
		stderr: &stderr,
		urlCh:  make(chan string, 1),
	}
	require.NoError(t, target.scanStderr(io.NopCloser(bytes.NewBufferString("Debugger listening on ws://127.0.0.1:1234/id\napplication error\n"))))

	require.Equal(t, "ws://127.0.0.1:1234/id", <-target.urlCh)
	require.Equal(t, "Debugger listening on ws://127.0.0.1:1234/id\napplication error\n", stderr.String())
}

func TestScanStderrForwardsLongLines(t *testing.T) {
	line := strings.Repeat("x", 1024*1024+1) + "\n"
	var stderr bytes.Buffer
	target := &Process{stderr: &stderr, urlCh: make(chan string, 1)}

	require.NoError(t, target.scanStderr(io.NopCloser(strings.NewReader(line))))
	require.Equal(t, len(line), stderr.Len())
}

func TestScanStderrReturnsForwardingError(t *testing.T) {
	want := errors.New("write failed")
	target := &Process{stderr: errorWriter{err: want}, urlCh: make(chan string, 1)}

	err := target.scanStderr(io.NopCloser(strings.NewReader("first\nsecond\n")))
	require.ErrorIs(t, err, want)
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestResolveEntryMain(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"main":"src/app.js"}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "src", "app.js"), []byte(""), 0o600))

	entry, err := resolveEntry(dir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "src", "app.js"), entry)
}

func TestResolveEntryBin(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"bin":{"app":"bin/cli.js"}}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "cli.js"), []byte(""), 0o600))

	entry, err := resolveEntry(dir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "bin", "cli.js"), entry)
}

func TestResolveEntryIndex(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.js"), []byte(""), 0o600))

	entry, err := resolveEntry(dir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "index.js"), entry)
}

func TestResolveEntryMissing(t *testing.T) {
	_, err := resolveEntry(t.TempDir())
	require.Error(t, err)
}
