package engine

import (
	"context"
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

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
func loadPackages(ctx context.Context, sourcePath string) ([]*packages.Package, error) {
	// Build config and loading packages
	pkgConfig := packages.Config{
		Mode:    packages.LoadAllSyntax,
		Context: ctx,
		Dir:     sourcePath,
	}

	// Pattern resolves to Config.Dir's source dir
	pkgs, err := packages.Load(&pkgConfig, CwdLimiter)
	if err != nil {
		return nil, err
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
