package analysis

import (
	"testing"

	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraph(t *testing.T) {
	t.Parallel()

	graph := BuildGraph()

	require.NotNil(t, graph.Nodes[graph.Root])
	assert.Equal(t, "redis-server", graph.Nodes[graph.Root].Name)

	names := CommandNames()
	require.NotEmpty(t, names)
	assert.Equal(t, "APPEND", names[0])
	assert.Equal(t, len(names)+1, len(graph.Nodes))

	for index, name := range names {
		id := model.NodeID(index + 2)
		node := graph.Nodes[id]
		require.NotNil(t, node, "missing node for %s", name)
		assert.Equal(t, name, node.Name)
		require.NotEqual(t, uint64(0), uint64(node.Syntax.File))
		file := graph.Files[node.Syntax.File]
		require.NotNil(t, file)
		assert.Equal(t, ModulePath+"/"+name, file.Path)
	}
}

func TestCommandNamesUniqueSorted(t *testing.T) {
	t.Parallel()

	names := CommandNames()
	seen := make(map[string]struct{}, len(names))
	for i, name := range names {
		if i > 0 {
			assert.True(t, names[i-1] < name, "not sorted at %q", name)
		}
		_, dup := seen[name]
		assert.False(t, dup, "duplicate command %q", name)
		seen[name] = struct{}{}
	}
	assert.Contains(t, names, "GET")
	assert.Contains(t, names, "SET")
	assert.Contains(t, names, "PING")
	assert.Contains(t, names, "INFO")
}
