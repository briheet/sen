// Package process owns a compiled Rust target process.
package process

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/briheet/sen/internal/adapters"
	rustbuild "github.com/briheet/sen/internal/adapters/rust/build"
)

// Process manages the target lifecycle without attaching its terminal input.
type Process struct {
	Prepared *rustbuild.Prepared
	cmd      *exec.Cmd
	started  chan struct{}
	exit     chan struct{}
	startErr error
}

// New creates the command for a prepared Cargo executable.
func New(prepared *rustbuild.Prepared, runArgs []string, output adapters.Output) *Process {
	cmd := exec.Command(prepared.Binary, runArgs...)
	cmd.Dir = prepared.SourceRoot
	cmd.Stdout = writer(output.Stdout)
	cmd.Stderr = writer(output.Stderr)
	cmd.Env = os.Environ()
	return &Process{Prepared: prepared, cmd: cmd, started: make(chan struct{}), exit: make(chan struct{})}
}

// AddEnv adds or replaces child-only environment variables before Start.
func (p *Process) AddEnv(values ...string) {
	for _, value := range values {
		key, _, ok := strings.Cut(value, "=")
		if !ok {
			continue
		}
		prefix := key + "="
		replaced := false
		for index := range p.cmd.Env {
			if strings.HasPrefix(p.cmd.Env[index], prefix) {
				p.cmd.Env[index], replaced = value, true
			}
		}
		if !replaced {
			p.cmd.Env = append(p.cmd.Env, value)
		}
	}
}

func writer(value io.Writer) io.Writer {
	if value == nil {
		return io.Discard
	}
	return value
}

// Start launches the target.
func (p *Process) Start() error { p.startErr = p.cmd.Start(); close(p.started); return p.startErr }

// PID returns the child process identifier.
func (p *Process) PID() int {
	if p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

// Wait blocks until the target exits.
func (p *Process) Wait() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	err := p.cmd.Wait()
	close(p.exit)
	return err
}

// Stop terminates the target gracefully, then kills it if necessary.
func (p *Process) Stop() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	if p.cmd.Process == nil {
		return nil
	}
	_ = p.cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-p.exit:
		return nil
	case <-time.After(2 * time.Second):
		if err := p.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		return nil
	}
}

// Cleanup removes the isolated Cargo build.
func (p *Process) Cleanup() error { return p.Prepared.Cleanup() }
