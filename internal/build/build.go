// Package build constructs and manages engines for configured services.
package build

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"

	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	runlog "github.com/briheet/sen/internal/log"
	"go.uber.org/zap"
)

var errNoRunnableServices = errors.New("no runnable services configured")

// Group owns the engines created for one sen run.
type Group struct {
	Engines []*engine.Engine

	logs   *runlog.Run
	logger *zap.Logger
}

// New constructs engines for supported services in configuration order.
func New(ctx context.Context, configuration *config.Config) (*Group, error) {
	runnable := false
	for _, service := range configuration.Services {
		if service.Type == config.ServiceTypeServer {
			runnable = true
			break
		}
	}
	if !runnable {
		return nil, errNoRunnableServices
	}

	logs, err := runlog.New(configuration.Project.Name)
	if err != nil {
		return nil, err
	}
	group := &Group{
		Engines: make([]*engine.Engine, 0, len(configuration.Services)),
		logs:    logs,
		logger:  logs.Logger().With(zap.String("component", "engine")),
	}
	group.logger.Info("run initialized", zap.String("log_path", logs.Path()))
	if path := logs.DebugPath(); path != "" {
		group.logger.Info("TUI debugging enabled", zap.String("debug_log_path", path))
	}

	for _, service := range configuration.Services {
		fields := serviceLogFields(service)
		if service.Type != config.ServiceTypeServer {
			group.logger.Info("supporting service loaded", fields...)
			continue
		}
		group.logger.Info("building service", fields...)
		target, err := engine.NewEngine(ctx, service, logs.Output(service.Name, string(service.Type), string(service.Lang)))
		if err != nil {
			group.logger.Error("build service failed", append(fields, zap.Error(err))...)
			cleanupErr := group.Cleanup()
			return nil, errors.Join(
				errors.New("build service "+strconv.Quote(service.Name)),
				err,
				errors.New("engine log "+strconv.Quote(logs.Path())),
				cleanupErr,
			)
		}
		group.Engines = append(group.Engines, target)
		group.logger.Info("service built", fields...)
	}
	return group, nil
}

func serviceLogFields(service config.Service) []zap.Field {
	fields := []zap.Field{
		zap.String("service", service.Name),
		zap.String("service_type", string(service.Type)),
	}
	if service.Lang != "" {
		fields = append(fields, zap.String("service_lang", string(service.Lang)))
	}
	if service.Provider != "" {
		fields = append(fields, zap.String("service_provider", string(service.Provider)))
	}
	return fields
}

// LogPath returns the engine log path.
func (g *Group) LogPath() string {
	return g.logs.Path()
}

// DebugWriter returns the TUI message log when DEBUG is set.
func (g *Group) DebugWriter() io.Writer {
	return g.logs.DebugWriter()
}

// DebugPath returns the TUI debug log path when debugging is enabled.
func (g *Group) DebugPath() string {
	return g.logs.DebugPath()
}

// Run starts every engine and waits for all services to exit.
func (g *Group) Run() error {
	type result struct {
		name string
		err  error
	}
	results := make(chan result, len(g.Engines))
	for _, target := range g.Engines {
		name := target.Service.Name
		go func() {
			g.logger.Info("service run started", zap.String("service", name))
			err := target.Run()
			if err != nil {
				g.logger.Error("service run failed", zap.String("service", name), zap.Error(err))
			} else {
				g.logger.Info("service run completed", zap.String("service", name))
			}
			results <- result{name: name, err: err}
		}()
	}
	var errs []error
	for range g.Engines {
		result := <-results
		if result.err != nil {
			errs = append(errs, errors.Join(errors.New("run service "+strconv.Quote(result.name)), result.err))
		}
	}
	return errors.Join(errs...)
}

// Stop terminates every running engine.
func (g *Group) Stop() error {
	var errs []error
	for _, target := range g.Engines {
		if err := target.Stop(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			name := target.Service.Name
			g.logger.Error("service stop failed", zap.String("service", name), zap.Error(err))
			errs = append(errs, errors.Join(errors.New("stop service "+strconv.Quote(name)), err))
		}
	}
	return errors.Join(errs...)
}

// Cleanup removes resources owned by every engine.
func (g *Group) Cleanup() error {
	var errs []error
	for _, target := range g.Engines {
		if err := target.Cleanup(); err != nil {
			name := target.Service.Name
			g.logger.Error("service cleanup failed", zap.String("service", name), zap.Error(err))
			errs = append(errs, errors.Join(errors.New("cleanup service "+strconv.Quote(name)), err))
		}
	}
	if g.logger != nil {
		g.logger.Info("run cleanup completed")
	}
	if g.logs != nil {
		if err := g.logs.Close(); err != nil {
			errs = append(errs, errors.Join(errors.New("close engine log"), err))
		}
	}
	return errors.Join(errs...)
}
