package runtime

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	tempDirPattern       = "senbon-*"
	binaryFileName       = "main"
	virtualFilePrefix    = "zz_"
	virtualFileExtension = ".go"
)

// Process handles the lifecycle of go's program
type Process struct {
	BinaryPath string
	RunCmd     *exec.Cmd
	tempDir    string
}

// This will build with overlay and store process context
func NewProcess(ctx context.Context, sourceDir string) (process *Process, err error) {
	// Create a tempDir for binary and overlay stuff
	tempDir, err := os.MkdirTemp("", tempDirPattern)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	// Create an overlay before build for collector
	overlayPath, err := CreateOverlay(ctx, sourceDir, tempDir)
	if err != nil {
		return nil, err
	}

	// Build process
	binaryPath := filepath.Join(tempDir, binaryFileName)
	buildCmd := exec.CommandContext(ctx, "go", "build", "-overlay="+overlayPath, "-o", binaryPath, ".")
	buildCmd.Dir = sourceDir
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdin = os.Stdin
	buildCmd.Stdout = os.Stdout
	if err := buildCmd.Run(); err != nil {
		return nil, err
	}

	runCmd := exec.CommandContext(ctx, binaryPath)
	runCmd.Dir = sourceDir
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin
	runCmd.Stdout = os.Stdout

	return &Process{BinaryPath: binaryPath, RunCmd: runCmd, tempDir: tempDir}, nil
}

// Run starts the built program and waits for it to exit.
func (p *Process) Run() error {
	return p.RunCmd.Run()
}

// Stop terminates the running program.
func (p *Process) Stop() error {
	return p.RunCmd.Process.Kill()
}

// Cleanup removes the temporary build directory.
func (p *Process) Cleanup() error {
	return os.RemoveAll(p.tempDir)
}
