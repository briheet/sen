package analysis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraph(t *testing.T) {
	t.Parallel()

	graph := BuildGraph(
		[]Statement{{QueryID: 111, Query: "SELECT 1"}, {QueryID: 222, Query: "INSERT INTO t"}},
		[]Table{{Name: "users"}, {Name: "orders"}},
	)

	require.NotNil(t, graph.Nodes[graph.Root])
	assert.Equal(t, "postgres-server", graph.Nodes[graph.Root].Name)

	assert.Equal(t, 5, len(graph.Nodes))

	names := make(map[string]bool)
	for _, node := range graph.Nodes {
		if node.ID == graph.Root {
			continue
		}
		names[node.Name] = true
		file := graph.Files[node.Syntax.File]
		require.NotNil(t, file)
		assert.Contains(t, file.Path, ModulePath+"/")
	}
	assert.True(t, names["SELECT 1"])
	assert.True(t, names["INSERT INTO t"])
	assert.True(t, names["users"])
	assert.True(t, names["orders"])
}

func TestPaths(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "sen/postgres/stmt/42", StmtPath(42))
	assert.Equal(t, "sen/postgres/table/users", TablePath("users"))
}

func TestBuildGraphEmpty(t *testing.T) {
	t.Parallel()

	graph := BuildGraph(nil, nil)
	require.Equal(t, 1, len(graph.Nodes))
	assert.Equal(t, "postgres-server", graph.Nodes[graph.Root].Name)
}

func TestStatementLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "SELECT 1", StatementLabel(" SELECT  1\n"))
	assert.Equal(t, strings.Repeat("x", maxLabel)+"…", StatementLabel(strings.Repeat("x", 100)))
}
