// Package golang integrates Go source analysis and runtime collection.
package golang

import (
	"context"
	"errors"
	"fmt"
	"go/types"
	"strings"

	"github.com/briheet/senbon/internal/adapters"
	"github.com/briheet/senbon/internal/adapters/golang/analysis"
	"github.com/briheet/senbon/internal/model"
	"golang.org/x/tools/go/packages"
)

const currentDirectory = "."

// Adapter analyzes Go applications.
type Adapter struct{}

var _ adapters.Analyzer = (*Adapter)(nil)

// Analyze loads and converts a Go application into the normalized graph.
func (*Adapter) Analyze(ctx context.Context, sourcePath string) (*model.StaticGraph, string, error) {
	packages, err := LoadPackages(ctx, sourcePath)
	if err != nil {
		return nil, "", err
	}
	graph, err := analysis.GetGraph(packages)
	if err != nil {
		return nil, "", err
	}
	return graph, packages[0].Module.Path, nil
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

// This function helps with setting required config, loading source info
func LoadPackages(ctx context.Context, sourcePath string) ([]*packages.Package, error) {
	// Build config and loading packages
	pkgConfig := packages.Config{
		Mode:    packages.LoadAllSyntax | packages.NeedModule,
		Context: ctx,
		Dir:     sourcePath,
	}

	// Pattern resolves to Config.Dir's source dir
	pkgs, err := packages.Load(&pkgConfig, currentDirectory)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 || pkgs[0].Module == nil {
		return nil, errors.New("source package has no module metadata")
	}

	// Temp struct for package loading errors
	var pkgErrors PackageErrors
	for _, pkg := range pkgs {
		// For any package, analyze Errors and Type Errors
		if len(pkg.Errors) > 0 || len(pkg.TypeErrors) > 0 {
			pkgErrors.Packages = append(pkgErrors.Packages, PackageError{
				Package:    pkg.PkgPath,
				Errors:     pkg.Errors,
				TypeErrors: pkg.TypeErrors,
			})
		}
	}

	// Return if errors were found out
	if len(pkgErrors.Packages) > 0 {
		return nil, &pkgErrors
	}

	return pkgs, nil
}
