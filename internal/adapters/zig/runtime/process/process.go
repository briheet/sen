// Package process builds and manages the target Zig program.
package process

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/briheet/senbon/internal/adapters/zig/analysis"
)

const (
	tempDirPattern = "senbon-zig-*"
	samplesEnv     = "SENBON_SAMPLES_FILE"
)

//go:embed sampler.zig
var samplerSource []byte

// Process handles the lifecycle of the target Zig program.
type Process struct {
	BinPath     string
	DsymPath    string
	SamplesFile string
	cmd         *exec.Cmd
	started     chan struct{}
	startErr    error
	exit        chan struct{}
	waitOnce    sync.Once
	tempDir     string
}

// NewProcess builds the target with the sampler wrapper injected.
func NewProcess(ctx context.Context, sourceDir string, project *analysis.Project) (*Process, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), tempDirPattern)
	if err != nil {
		return nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	samplerPath, err := writeFile(tempDir, "sampler.zig", samplerSource)
	if err != nil {
		cleanup()
		return nil, err
	}
	binPath := filepath.Join(tempDir, "main")
	build := exec.CommandContext(ctx, "zig", buildArgs(project, sourceDir, samplerPath, binPath, tempDir)...)
	build.Dir = sourceDir
	if output, err := build.CombinedOutput(); err != nil {
		cleanup()
		return nil, &BuildError{Output: string(output), Err: err}
	}

	cmd := exec.CommandContext(ctx, binPath)
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), samplesEnv+"="+filepath.Join(tempDir, "samples.txt"))

	return &Process{
		BinPath:     binPath,
		DsymPath:    dsymPath(binPath),
		SamplesFile: filepath.Join(tempDir, "samples.txt"),
		cmd:         cmd,
		started:     make(chan struct{}),
		exit:        make(chan struct{}),
		tempDir:     tempDir,
	}, nil
}

// dsymPath extracts DWARF into a dSYM bundle on macOS.
func dsymPath(binPath string) string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	if _, err := exec.LookPath("dsymutil"); err != nil {
		return ""
	}
	output, err := exec.Command("dsymutil", binPath).CombinedOutput()
	if err != nil {
		return ""
	}
	_ = output
	return binPath + ".dSYM/Contents/Resources/DWARF/" + filepath.Base(binPath)
}

// buildArgs assembles the module graph for zig build-exe.
func buildArgs(project *analysis.Project, sourceDir, samplerPath, binPath, cacheDir string) []string {
	args := []string{"build-exe", "-lc"}

	args = append(args, "--dep", "user=user", "-Mroot="+samplerPath)
	args = append(args, depFlags(project, project.Entry)...)
	args = append(args, "-Muser="+project.Entry)

	seen := map[string]bool{project.Entry: true}
	for name, file := range project.Modules {
		if seen[file] {
			continue
		}
		seen[file] = true
		args = append(args, depFlags(project, file)...)
		args = append(args, "-M"+name+"="+file)
	}
	return append(args, "-femit-bin="+binPath, "-O", "Debug", "--cache-dir", cacheDir)
}

// depFlags returns the --dep entries for one module's local imports.
func depFlags(project *analysis.Project, file string) []string {
	var flags []string
	for _, name := range project.Imports[file] {
		flags = append(flags, "--dep", name+"="+name)
	}
	return flags
}

// BuildError reports a failed target build.
type BuildError struct {
	Output string
	Err    error
}

func (e *BuildError) Error() string {
	return "zig build failed: " + e.Output + e.Err.Error()
}

func (e *BuildError) Unwrap() error { return e.Err }

// Start launches the target.
func (p *Process) Start() error {
	if err := p.cmd.Start(); err != nil {
		p.startErr = err
		close(p.started)
		return err
	}
	close(p.started)
	return nil
}

// Wait blocks until the target exits.
func (p *Process) Wait() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	err := p.cmd.Wait()
	p.waitOnce.Do(func() { close(p.exit) })
	return err
}

// Stop terminates the target gracefully, escalating to a kill.
func (p *Process) Stop() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Signal(syscall.SIGTERM)
	}
	select {
	case <-p.exit:
	case <-time.After(2 * time.Second):
		if p.cmd.Process != nil {
			_ = p.cmd.Process.Kill()
		}
	}
	return nil
}

// Cleanup removes the temporary build directory.
func (p *Process) Cleanup() error {
	return os.RemoveAll(p.tempDir)
}

// writeFile writes source to the temporary directory.
func writeFile(tempDir, name string, source []byte) (string, error) {
	path := filepath.Join(tempDir, name)
	if err := os.WriteFile(path, source, 0o600); err != nil {
		return "", err
	}
	return path, nil
}
