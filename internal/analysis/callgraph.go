package analysis

import (
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/ssa"
)

// This functions helps in building callgraph
func BuildCallgraph(mainFunc *ssa.Function) *rta.Result {
	result := rta.Analyze(
		[]*ssa.Function{mainFunc},
		true,
	)
	return result
}
