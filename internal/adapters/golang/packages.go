// Package golang integrates Go source analysis and runtime collection.
package golang

import (
	"context"
	"errors"
	"fmt"
	"go/types"
	"slices"
	"strings"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/golang/analysis"
	targetruntime "github.com/briheet/sen/internal/adapters/golang/runtime"
	"github.com/briheet/sen/internal/model"
	"golang.org/x/tools/go/packages"
)

const currentDirectory = "."

// Adapter analyzes Go applications.
type Adapter struct{}

var _ adapters.Application = (*Adapter)(nil)

// Analyze loads and converts a Go application into the normalized graph.
func (*Adapter) Analyze(ctx context.Context, sourcePath string, buildArgs []string) (*model.StaticGraph, string, error) {
	packages, err := LoadPackages(ctx, sourcePath, buildArgs)
	if err != nil {
		return nil, "", err
	}
	graph, err := analysis.GetGraph(packages)
	if err != nil {
		return nil, "", err
	}
	return graph, packages[0].Module.Path, nil
}

// Open builds the instrumented Go target.
func (*Adapter) Open(ctx context.Context, sourcePath string, buildArgs, runArgs []string, output adapters.Output) (adapters.Runtime, error) {
	return targetruntime.NewRuntime(ctx, sourcePath, buildArgs, runArgs, output)
}

// Go packages error
type PackageError struct {
	Package    string
	Errors     []packages.Error
	TypeErrors []types.Error
}

// Consolidate errors
type PackageErrors struct {
	Packages []PackageError
}

// Implements the Error interface
func (e *PackageErrors) Error() string {
	var b strings.Builder

	for _, pkg := range e.Packages {
		fmt.Fprintf(&b, "package %s has errors:\n", pkg.Package)

		for _, err := range pkg.Errors {
			fmt.Fprintf(&b, "  %s\n", err)
		}

		for _, err := range pkg.TypeErrors {
			fmt.Fprintf(&b, "  %s\n", err)
		}
	}

	return b.String()
}

// LoadPackages loads typed syntax for the service's local import closure.
func LoadPackages(ctx context.Context, sourcePath string, buildArgs []string) ([]*packages.Package, error) {
	// Discover the local import closure without parsing dependency source.
	metadataConfig := packages.Config{
		Mode:       packages.LoadImports | packages.NeedDeps | packages.NeedModule,
		Context:    ctx,
		Dir:        sourcePath,
		BuildFlags: buildArgs,
	}
	metadata, err := packages.Load(&metadataConfig, currentDirectory)
	if err != nil {
		return nil, err
	}
	if len(metadata) == 0 || metadata[0].Module == nil {
		return nil, errors.New("source package has no module metadata")
	}

	rootPath := metadata[0].PkgPath
	modulePath := metadata[0].Module.Path
	paths := make([]string, 0, len(metadata))
	packages.Visit(metadata, nil, func(pkg *packages.Package) {
		if pkg.Module != nil && pkg.Module.Path == modulePath {
			paths = append(paths, pkg.PkgPath)
		}
	})
	slices.Sort(paths)
	paths = slices.Compact(paths)

	// Parse and type-check project packages; imports use compiler export data.
	syntaxConfig := packages.Config{
		Mode:       packages.LoadSyntax | packages.NeedModule,
		Context:    ctx,
		Dir:        sourcePath,
		BuildFlags: buildArgs,
	}
	pkgs, err := packages.Load(&syntaxConfig, paths...)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, errors.New("source package was not loaded")
	}
	slices.SortFunc(pkgs, func(left, right *packages.Package) int {
		if left.PkgPath == rootPath && right.PkgPath != rootPath {
			return -1
		}
		if right.PkgPath == rootPath && left.PkgPath != rootPath {
			return 1
		}
		return strings.Compare(left.PkgPath, right.PkgPath)
	})

	var pkgErrors PackageErrors
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 || len(pkg.TypeErrors) > 0 {
			pkgErrors.Packages = append(pkgErrors.Packages, PackageError{
				Package:    pkg.PkgPath,
				Errors:     pkg.Errors,
				TypeErrors: pkg.TypeErrors,
			})
		}
	}

	if len(pkgErrors.Packages) > 0 {
		return nil, &pkgErrors
	}
	for _, pkg := range pkgs {
		if pkg.Types == nil || pkg.IllTyped {
			return nil, errors.New("source package is not well typed")
		}
	}

	return pkgs, nil
}
