package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProcessWithOverlay(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example\n\ngo 1.25\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	process, err := NewProcess(context.Background(), sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = process.Cleanup() })
	if _, err := os.Stat(process.BinaryPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(process.tempDir, OverlayFileName)); err != nil {
		t.Fatal(err)
	}
	if err := process.Run(); err != nil {
		t.Fatal(err)
	}
}
