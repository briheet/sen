// Package log owns structured logs for one sen run.
package log

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

var errInvalidProjectName = errors.New("project name must not be a path")

const (
	applicationName = "sen"
	engineLogName   = "engine.log"
	timestampFormat = "20060102T150405.000000000Z"
)

// Run owns the logger and process streams for one invocation.
type Run struct {
	dir     string
	path    string
	file    *os.File
	logger  *zap.Logger
	writers []*zapio.Writer
	closed  bool
}

// New creates a project-scoped log directory in the user cache.
func New(project string) (*Run, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, errors.Join(errors.New("resolve user cache directory"), err)
	}
	return newRun(cacheDir, project, time.Now().UTC())
}

func newRun(cacheDir, project string, startedAt time.Time) (*Run, error) {
	if project == "" || project == "." || project == ".." || filepath.Base(project) != project {
		return nil, errInvalidProjectName
	}
	baseDir := filepath.Join(cacheDir, applicationName)
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return nil, errors.Join(errors.New("create sen cache directory"), err)
	}
	dir := filepath.Join(baseDir, project+"-"+startedAt.Format(timestampFormat))
	if err := os.Mkdir(dir, 0o700); err != nil {
		return nil, errors.Join(errors.New("create run log directory"), err)
	}
	path := filepath.Join(dir, engineLogName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		_ = os.Remove(dir)
		return nil, errors.Join(errors.New("create engine log"), err)
	}

	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	sink := zapcore.Lock(zapcore.AddSync(file))
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoder), sink, zap.InfoLevel)).With(
		zap.String("project", project),
	)
	return &Run{dir: dir, path: path, file: file, logger: logger}, nil
}

// Logger returns the run's structured application logger.
func (r *Run) Logger() *zap.Logger {
	return r.logger
}

// Dir returns the directory reserved for this run's logs.
func (r *Run) Dir() string {
	return r.dir
}

// Path returns the engine log path.
func (r *Run) Path() string {
	return r.path
}

// Output creates tagged stdout and stderr streams for a service.
func (r *Run) Output(name, serviceType, serviceLang string) adapters.Output {
	logger := r.logger.With(
		zap.String("component", "service"),
		zap.String("service", name),
		zap.String("service_type", serviceType),
		zap.String("service_lang", serviceLang),
	)
	stdout := &zapio.Writer{Log: logger.With(zap.String("stream", "stdout")), Level: zap.InfoLevel}
	stderr := &zapio.Writer{Log: logger.With(zap.String("stream", "stderr")), Level: zap.ErrorLevel}
	r.writers = append(r.writers, stdout, stderr)
	return adapters.Output{Stdout: stdout, Stderr: stderr}
}

// Close flushes process streams and closes the run log.
func (r *Run) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	var errs []error
	for _, writer := range r.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := r.logger.Sync(); err != nil {
		errs = append(errs, err)
	}
	if err := r.file.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
