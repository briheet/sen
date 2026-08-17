package process

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

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
