// Package process builds and manages the target Go program.
package process

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/briheet/sen/internal/adapters"
)

const (
	tempDirPattern           = "sen-*"
	binaryFileName           = "main"
	collectorSocketExtension = ".sock"
	collectorSocketEnvName   = "SEN_COLLECTOR_SOCKET"
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

// NewProcess builds the target program with the sen overlay.
func NewProcess(ctx context.Context, sourceDir string, buildArgs, runArgs []string, output adapters.Output) (process *Process, err error) {
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
	args := append([]string{"build"}, buildArgs...)
	args = append(args, "-overlay="+overlayPath, "-o", binaryPath, ".")
	buildCmd := exec.CommandContext(ctx, "go", args...)
	buildCmd.Dir = sourceDir
	buildCmd.Stderr = output.Stderr
	buildCmd.Stdin = os.Stdin
	buildCmd.Stdout = output.Stdout
	if err := buildCmd.Run(); err != nil {
		return nil, err
	}

	runCmd := exec.CommandContext(ctx, binaryPath, runArgs...)
	runCmd.Dir = sourceDir
	runCmd.Stderr = output.Stderr
	// Bubble Tea owns terminal input while services run.
	runCmd.Stdout = output.Stdout
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

// Start launches the target program.
func (p *Process) Start() error {
	p.startErr = p.RunCmd.Start()
	close(p.started)
	return p.startErr
}

// PID returns the running target's process identifier.
func (p *Process) PID() int {
	if p.RunCmd.Process == nil {
		return 0
	}
	return p.RunCmd.Process.Pid
}

// Wait blocks until the target program exits.
func (p *Process) Wait() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	return p.RunCmd.Wait()
}

// Run starts the target program and waits for it to exit.
func (p *Process) Run() error {
	if err := p.Start(); err != nil {
		return err
	}
	return p.Wait()
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
