// Package process builds and manages the target OCaml program.
package process

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const (
	tempDirPattern = "senbon-ocaml-*"
	eventsDirEnv   = "OCAML_RUNTIME_EVENTS_DIR"
	eventsStartEnv = "OCAML_RUNTIME_EVENTS_START"
	eventsPreserve = "OCAML_RUNTIME_EVENTS_PRESERVE"
)

//go:embed instrument.ml
var instrumentSource []byte

//go:embed profiler.ml
var profilerSource []byte

// Process handles the lifecycle of the target OCaml program.
type Process struct {
	BinPath     string
	EventsDir   string
	FunctionMap string // JSON id -> function name
	cmd         *exec.Cmd
	started     chan struct{}
	startErr    error
	exit        chan struct{}
	waitOnce    sync.Once
	tempDir     string
}

// NewProcess instruments and builds the target program with ocamlopt +
// runtime_events, then wraps each function body with the span emitter.
func NewProcess(ctx context.Context, sourceDir, entry string) (*Process, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), tempDirPattern)
	if err != nil {
		return nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	// 1. write embedded sources into the temp build dir
	instrumentPath := filepath.Join(tempDir, "instrument.ml")
	profilerPath := filepath.Join(tempDir, "profiler.ml")
	if err := writeFile(instrumentPath, instrumentSource); err != nil {
		cleanup()
		return nil, err
	}
	if err := writeFile(profilerPath, profilerSource); err != nil {
		cleanup()
		return nil, err
	}

	// 2. compile instrument.ml to a helper binary (uses compiler-libs)
	instrumentBin := filepath.Join(tempDir, "instrument")
	if _, err := run(ctx, tempDir, "ocamlc", "-I", ocamlCompileInclude(), ocamlCommon(),
		filepath.Base(instrumentPath), "-o", instrumentBin); err != nil {
		cleanup()
		return nil, err
	}
	defer func() { _ = os.Remove(instrumentBin) }()

	// 3. transform the entry source and produce the id->name map
	instrumented := filepath.Join(tempDir, "main.ml")
	functionMap := filepath.Join(tempDir, "fns.json")
	if _, err := run(ctx, sourceDir, instrumentBin, entry, instrumented, functionMap); err != nil {
		cleanup()
		return nil, err
	}

	// 4. build the instrumented target with profiler + runtime_events
	binPath := filepath.Join(tempDir, "main")
	link, err := run(ctx, tempDir, "ocamlopt", "-I", ocamlCompileInclude(),
		"-I", ocamlRuntimeEventsInclude(), ocamlRuntimeEvents(),
		filepath.Base(profilerPath), filepath.Base(instrumented), "-o", binPath)
	if err != nil {
		cleanup()
		return nil, &BuildError{Output: string(link), Err: err}
	}

	// 5. runtime-events directory + launch env
	eventsDir := filepath.Join(tempDir, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		cleanup()
		return nil, err
	}

	cmd := exec.CommandContext(ctx, binPath)
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		eventsDirEnv+"="+eventsDir,
		eventsStartEnv+"=1",
		eventsPreserve+"=1",
	)

	return &Process{
		BinPath:     binPath,
		EventsDir:   eventsDir,
		FunctionMap: functionMap,
		cmd:         cmd,
		started:     make(chan struct{}),
		exit:        make(chan struct{}),
		tempDir:     tempDir,
	}, nil
}

// writeFile writes bytes to a path with 0o600 perms.
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

// run executes a command in dir and returns its combined output.
func run(ctx context.Context, dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// ocamlCompileInclude returns <stdlib>/compiler-libs.
func ocamlCompileInclude() string {
	return ocamlStdlib() + "/compiler-libs"
}

// ocamlRuntimeEventsInclude returns <stdlib>/runtime_events.
func ocamlRuntimeEventsInclude() string {
	return ocamlStdlib() + "/runtime_events"
}

// ocamlCommon returns the compiler-libs common archive.
func ocamlCommon() string {
	return ocamlStdlib() + "/compiler-libs/ocamlcommon.cma"
}

// ocamlRuntimeEvents returns the runtime_events archive.
func ocamlRuntimeEvents() string {
	return ocamlStdlib() + "/runtime_events/runtime_events.cmxa"
}

var ocamlStdlibPath = ""

// ocamlStdlib caches `ocamlc -where`.
func ocamlStdlib() string {
	if ocamlStdlibPath != "" {
		return ocamlStdlibPath
	}
	if output, err := exec.Command("ocamlc", "-where").Output(); err == nil {
		ocamlStdlibPath = string(output[:len(output)-1])
	}
	return ocamlStdlibPath
}

// BuildError reports a failed target build.
type BuildError struct {
	Output string
	Err    error
}

func (e *BuildError) Error() string { return "ocaml build failed: " + e.Output + e.Err.Error() }
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
		_ = p.cmd.Process.Kill()
	}
	select {
	case <-p.exit:
	case <-time.After(2 * time.Second):
	}
	return nil
}

// Cleanup removes the temporary build directory.
func (p *Process) Cleanup() error {
	return os.RemoveAll(p.tempDir)
}

// EventsFile returns the runtime events ring buffer for the given pid.
func (p *Process) EventsFile(pid int) string {
	return filepath.Join(p.EventsDir, itoa(pid)+".events")
}

// PID returns the target process id, or zero if not started.
func (p *Process) PID() int {
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
