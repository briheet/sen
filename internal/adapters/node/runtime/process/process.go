// Package process builds and manages the target Node.js program.
package process

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/briheet/sen/internal/adapters"
)

const (
	tempDirPattern    = "sen-node-*"
	metricsFileEnv    = "SEN_METRICS_FILE"
	metricsIntervalMs = "SEN_METRICS_INTERVAL_MS"
	inspectAddr       = "127.0.0.1:0"
)

//go:embed shim.cjs
var shimSource []byte

// Process handles the lifecycle of the target Node.js program.
type Process struct {
	MetricsFile string
	cmd         *exec.Cmd
	stderr      io.Writer
	urlCh       chan string
	started     chan struct{}
	startErr    error
	exit        chan struct{}
	stderrErr   chan error
	waitOnce    sync.Once
	tempDir     string
}

// NewProcess builds the run command for the target program.
func NewProcess(ctx context.Context, sourceDir string, buildArgs, runArgs []string, output adapters.Output) (process *Process, err error) {
	sourceDir, err = filepath.Abs(sourceDir)
	if err != nil {
		return nil, err
	}
	sourceDir, err = filepath.EvalSymlinks(sourceDir)
	if err != nil {
		return nil, err
	}
	entry, err := resolveEntry(sourceDir)
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
	shimPath, err := writeShim(tempDir)
	if err != nil {
		return nil, err
	}
	metricsFile := filepath.Join(tempDir, "metrics.ndjson")

	args := []string{
		"--inspect=" + inspectAddr,
		"--require", shimPath,
	}
	// Node has no build phase; build arguments are runtime flags.
	args = append(args, buildArgs...)
	args = append(args, entry)
	args = append(args, runArgs...)
	cmd := exec.CommandContext(ctx, "node", args...)
	cmd.Dir = sourceDir
	cmd.Stdout = output.Stdout
	cmd.Env = append(os.Environ(),
		metricsFileEnv+"="+metricsFile,
		metricsIntervalMs+"=100",
	)

	stderr := output.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	return &Process{
		MetricsFile: metricsFile,
		cmd:         cmd,
		stderr:      stderr,
		urlCh:       make(chan string, 1),
		started:     make(chan struct{}),
		exit:        make(chan struct{}),
		stderrErr:   make(chan error, 1),
		tempDir:     tempDir,
	}, nil
}

// Start launches the target and begins capturing its inspector URL.
func (p *Process) Start() error {
	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := p.cmd.Start(); err != nil {
		p.startErr = err
		close(p.started)
		return err
	}
	close(p.started)
	go func() {
		p.stderrErr <- p.scanStderr(stderr)
	}()
	return nil
}

// WaitURL blocks until the inspector websocket URL is available.
func (p *Process) WaitURL(ctx context.Context) (string, error) {
	select {
	case url := <-p.urlCh:
		return url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Wait blocks until the target process exits.
func (p *Process) Wait() error {
	<-p.started
	if p.startErr != nil {
		return p.startErr
	}
	err := errors.Join(p.cmd.Wait(), <-p.stderrErr)
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

// scanStderr forwards target output and extracts the inspector URL.
func (p *Process) scanStderr(stderr io.ReadCloser) error {
	reader := bufio.NewReader(stderr)
	var forwardErr error
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			if index := strings.Index(line, "ws://"); index >= 0 {
				select {
				case p.urlCh <- strings.TrimSpace(line[index:]):
				default:
				}
			}
			if forwardErr == nil {
				_, forwardErr = io.WriteString(p.stderr, line)
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return forwardErr
			}
			return errors.Join(forwardErr, err)
		}
	}
}

// resolveEntry finds the target program's entry file.
func resolveEntry(sourceDir string) (string, error) {
	pkgPath := filepath.Join(sourceDir, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Main string          `json:"main"`
			Bin  json.RawMessage `json:"bin"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			if entry, ok := entryFile(sourceDir, pkg.Main); ok {
				return entry, nil
			}
			var single string
			if json.Unmarshal(pkg.Bin, &single) == nil && single != "" {
				if entry, ok := entryFile(sourceDir, single); ok {
					return entry, nil
				}
			}
			var multiple map[string]string
			if json.Unmarshal(pkg.Bin, &multiple) == nil {
				for _, binary := range multiple {
					if entry, ok := entryFile(sourceDir, binary); ok {
						return entry, nil
					}
				}
			}
		}
	}
	for _, name := range []string{"index.js", "index.mjs", "index.cjs", "index.ts"} {
		if entry, ok := entryFile(sourceDir, name); ok {
			return entry, nil
		}
	}
	return "", errors.New("no entry file found (package.json main/bin or index.*)")
}

func entryFile(sourceDir, path string) (string, bool) {
	if path == "" {
		return "", false
	}
	entry := path
	if !filepath.IsAbs(entry) {
		entry = filepath.Join(sourceDir, entry)
	}
	entry = filepath.Clean(entry)
	if _, err := os.Stat(entry); err != nil {
		return "", false
	}
	return entry, true
}

// writeShim writes the embedded metrics shim to the temporary directory.
func writeShim(tempDir string) (string, error) {
	path := filepath.Join(tempDir, "shim.cjs")
	if err := os.WriteFile(path, shimSource, 0o600); err != nil {
		return "", err
	}
	return path, nil
}
