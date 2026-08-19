package process

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/briheet/sen/internal/adapters"
)

func TestNewProcessWithOverlay(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example\n\ngo 1.25\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("//go:build configured\n\npackage main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	runArgs := []string{"--port", "8080"}
	var stdout, stderr bytes.Buffer
	process, err := NewProcess(
		context.Background(),
		sourceDir,
		[]string{"-tags=configured"},
		runArgs,
		adapters.Output{Stdout: &stdout, Stderr: &stderr},
	)
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
	if !slices.Equal(process.RunCmd.Args[1:], runArgs) {
		t.Fatalf("run arguments = %v, want %v", process.RunCmd.Args[1:], runArgs)
	}
	if process.RunCmd.Stdout != &stdout || process.RunCmd.Stderr != &stderr {
		t.Fatal("run command does not use configured output")
	}
	if process.RunCmd.Stdin != nil {
		t.Fatal("run command must not share TUI input")
	}
	if err := process.Run(); err != nil {
		t.Fatal(err)
	}
}
