package analysis

import (
	"errors"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const (
	MainFunction = "main"
)

var (
	ErrNoMainPackage  = errors.New("no main package found")
	ErrNoMainFunction = errors.New("no main function was found")
)

// This functions helps in building ssa packages and finding main and return it
func BuildPackagesAndReturnMain(pkgs []*packages.Package) (*ssa.Function, error) {
	// build ssa packages
	prog, ssaPkgs := ssautil.AllPackages(pkgs, ssa.InstantiateGenerics)
	prog.Build()

	// find main package if there
	mainPkg := ssautil.MainPackages(ssaPkgs)
	if len(mainPkg) == 0 {
		return nil, ErrNoMainPackage
	}

	// Return mainfunc if there
	for _, pkg := range mainPkg {
		mainF := pkg.Func(MainFunction)
		// Check nil
		if mainF != nil {
			return mainF, nil
		}
	}
	return nil, ErrNoMainFunction
}
