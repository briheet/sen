package process

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateOverlayAddsCollector(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example\n\ngo 1.25\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	overlayPath, err := CreateOverlay(sourceDir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "list", "-overlay="+overlayPath, "-json", ".")
	command.Dir = sourceDir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), virtualFilePrefix) {
		t.Fatalf("collector missing from package:\n%s", output)
	}
}
