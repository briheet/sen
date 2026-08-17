// Package ocaml integrates OCaml source analysis and runtime collection.
package ocaml

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/ocaml/analysis"
	"github.com/briheet/senbon/internal/adapters/ocaml/runtime"
	"github.com/briheet/senbon/internal/model"
)

// Adapter analyzes OCaml applications.
type Adapter struct {
	project *analysis.Project
}

var _ adapters.Application = (*Adapter)(nil)

// Analyze resolves the entry file and builds the normalized graph.
func (a *Adapter) Analyze(ctx context.Context, sourcePath string) (*model.StaticGraph, string, error) {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, "", err
	}
	entry, err := resolveEntry(sourcePath)
	if err != nil {
		return nil, "", err
	}
	project, err := analysis.Analyze(ctx, entry)
	if err != nil {
		return nil, "", err
	}
	a.project = project
	namespace := filepath.Base(filepath.Dir(entry))
	return project.Graph, namespace, nil
}

// Open builds the instrumented OCaml target.
func (a *Adapter) Open(ctx context.Context, sourcePath string) (adapters.Runtime, error) {
	if a.project == nil {
		return nil, errors.New("ocaml: analyze the project before opening")
	}
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}
	return runtime.NewRuntime(ctx, sourcePath, a.project.Entry)
}

// resolveEntry finds the main OCaml source file in the project directory.
func resolveEntry(sourcePath string) (string, error) {
	dir := sourcePath
	if entry, err := os.Stat(sourcePath); err == nil && !entry.IsDir() {
		return sourcePath, nil
	}
	// Prefer main.ml, else the single .ml file.
	if main := filepath.Join(dir, "main.ml"); fileExists(main) {
		return main, nil
	}
	var single string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".ml" && e.Name() != "main.ml" {
			if single != "" {
				return "", errors.New("ocaml: ambiguous entry; add a main.ml")
			}
			single = filepath.Join(dir, e.Name())
		}
	}
	if single == "" {
		return "", errors.New("ocaml: no .ml entry found")
	}
	return single, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
