package analysis

import (
	"testing"

	"github.com/briheet/senbon/internal/model"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	data := []byte(`{
		"entry": "/proj/calc.ml",
		"functions": ["add", "+", "fib", "-", "run"],
		"edges": [
			{"from": 4, "to": 0},
			{"from": 4, "to": 2}
		]
	}`)

	project, err := parse(data, "/proj/calc.ml")
	require.NoError(t, err)
	require.Equal(t, "/proj/calc.ml", project.Entry)
	require.Len(t, project.Graph.Nodes, 5)
	require.Len(t, project.Graph.Files, 1)

	// id 4 (run) -> node 5
	run := project.Graph.Nodes[5]
	require.Equal(t, "run", run.Name)
	require.Equal(t, []model.NodeID{1, 3}, run.Out) // add(0)->1, fib(2)->3

	// id 2 (fib) -> node 3
	fib := project.Graph.Nodes[3]
	require.Equal(t, "fib", fib.Name)
	require.Empty(t, fib.Out)

	// id 0 (add) -> node 1
	require.Equal(t, "add", project.Graph.Nodes[1].Name)

	// root is node 1 (the first binding), since the analyzer lists add first
	require.Equal(t, model.NodeID(1), project.Graph.Root)
}
