package analysis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildGraph(t *testing.T) {
	t.Parallel()
	graph := BuildGraph([]string{"127.0.0.1:3000", "127.0.0.1:3001"})
	require.Equal(t, 1+len(Operations)+2, len(graph.Nodes))
	require.Equal(t, "tigerbeetle-cluster", graph.Nodes[graph.Root].Name)
	require.Len(t, graph.Nodes[graph.Root].Out, len(Operations)+2)

	paths := make(map[string]bool)
	for _, file := range graph.Files {
		paths[file.Path] = true
	}
	require.True(t, paths[OperationPath("create_accounts")])
	require.True(t, paths[ReplicaPath(1)])
}

func TestNormalizeOperation(t *testing.T) {
	t.Parallel()
	operation, ok := NormalizeOperation("Operation.CREATE_ACCOUNTS")
	require.True(t, ok)
	require.Equal(t, "create_accounts", operation)
	_, ok = NormalizeOperation("pulse")
	require.False(t, ok)
}
