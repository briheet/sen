package log

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRunWritesStructuredServiceLogs(t *testing.T) {
	cacheDir := t.TempDir()
	startedAt := time.Date(2026, time.August, 18, 10, 20, 30, 123, time.UTC)
	run, err := newRun(cacheDir, "my-backend", startedAt, false)
	require.NoError(t, err)

	wantDir := filepath.Join(cacheDir, applicationName, "my-backend-20260818T102030.000000123Z")
	require.Equal(t, wantDir, run.Dir())
	require.Equal(t, filepath.Join(wantDir, engineLogName), run.Path())
	dirInfo, err := os.Stat(run.Dir())
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
	fileInfo, err := os.Stat(run.Path())
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

	run.Logger().Info("engine ready", zap.String("component", "engine"))
	output := run.Output("api", "server", "go")
	_, err = output.Stdout.Write([]byte("listening\n"))
	require.NoError(t, err)
	_, err = output.Stderr.Write([]byte("failed"))
	require.NoError(t, err)
	require.NoError(t, run.Close())
	require.NoError(t, run.Close())

	entries := readEntries(t, run.Path())
	require.Len(t, entries, 3)
	require.Equal(t, "my-backend", entries[0]["project"])
	require.Equal(t, "engine ready", entries[0]["msg"])
	require.Equal(t, "listening", entries[1]["msg"])
	require.Equal(t, "api", entries[1]["service"])
	require.Equal(t, "server", entries[1]["service_type"])
	require.Equal(t, "go", entries[1]["service_lang"])
	require.Equal(t, "stdout", entries[1]["stream"])
	require.Equal(t, "info", entries[1]["level"])
	require.Equal(t, "failed", entries[2]["msg"])
	require.Equal(t, "stderr", entries[2]["stream"])
	require.Equal(t, "error", entries[2]["level"])
}

func TestRunRejectsProjectPath(t *testing.T) {
	_, err := newRun(t.TempDir(), "../project", time.Now(), false)
	require.ErrorIs(t, err, errInvalidProjectName)
}

func TestRunCreatesTUIDebugLog(t *testing.T) {
	run, err := newRun(t.TempDir(), "my-backend", time.Now(), true)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(run.Dir(), tuiLogName), run.DebugPath())

	_, err = run.DebugWriter().Write([]byte("message\n"))
	require.NoError(t, err)
	require.NoError(t, run.Close())

	content, err := os.ReadFile(run.DebugPath())
	require.NoError(t, err)
	require.Equal(t, "message\n", string(content))
}

func readEntries(t *testing.T, path string) []map[string]any {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() { require.NoError(t, file.Close()) }()

	var entries []map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry map[string]any
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &entry))
		entries = append(entries, entry)
	}
	require.NoError(t, scanner.Err())
	return entries
}
