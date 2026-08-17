// Package process builds and manages the target Go program.
package process

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	tempDirPattern           = "senbon-*"
	binaryFileName           = "main"
	collectorSocketExtension = ".sock"
	collectorSocketEnvName   = "SENBON_COLLECTOR_SOCKET"
)

// Process handles the lifecycle of the target program.
type Process struct {
	BinaryPath      string
	CollectorSocket string
	RunCmd          *exec.Cmd
	tempDir         string
	started         chan struct{}
	startErr        error
}

// NewProcess builds the target program with the Senbon overlay.
func NewProcess(ctx context.Context, sourceDir string) (process *Process, err error) {
	sourceDir, err = filepath.Abs(sourceDir)
	if err != nil {
		return nil, err
	}
	sourceDir, err = filepath.EvalSymlinks(sourceDir)
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), tempDirPattern)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	overlayPath, err := CreateOverlay(sourceDir, tempDir)
	if err != nil {
		return nil, err
	}

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
	collectorSocket := filepath.Join(os.TempDir(), filepath.Base(tempDir)+collectorSocketExtension)
	runCmd.Env = append(os.Environ(), collectorSocketEnvName+"="+collectorSocket)

	return &Process{
		BinaryPath:      binaryPath,
		CollectorSocket: collectorSocket,
		RunCmd:          runCmd,
		tempDir:         tempDir,
		started:         make(chan struct{}),
	}, nil
}

// Run starts the target program and waits for it to exit.
func (p *Process) Run() error {
	p.startErr = p.RunCmd.Start()
	close(p.started)
	if p.startErr != nil {
		return p.startErr
	}
	return p.RunCmd.Wait()
}

// Stop terminates the running target program.
func (p *Process) Stop() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	return p.RunCmd.Process.Kill()
}

// Cleanup removes the temporary build directory.
func (p *Process) Cleanup() error {
	_ = os.Remove(p.CollectorSocket)
	return os.RemoveAll(p.tempDir)
}
