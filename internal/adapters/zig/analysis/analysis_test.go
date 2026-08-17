package analysis

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/briheet/senbon/internal/model"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	data := []byte(`{
		"root": 2,
		"entry": "/proj/src/main.zig",
		"files": [
			{"path": "/proj/src/services.zig", "imports": [], "functions": [
				{"id": 0, "name": "fib", "startLine": 0, "startCol": 4, "endLine": 3, "endCol": 1}
			]},
			{"path": "/proj/src/main.zig", "imports": ["std", "router", "services"], "functions": [
				{"id": 1, "name": "helper", "startLine": 0, "startCol": 4, "endLine": 2, "endCol": 1},
				{"id": 2, "name": "main", "startLine": 4, "startCol": 4, "endLine": 9, "endCol": 1}
			]},
			{"path": "/proj/src/router.zig", "imports": ["services"], "functions": [
				{"id": 3, "name": "route", "startLine": 2, "startCol": 4, "endLine": 7, "endCol": 1}
			]}
		],
		"edges": [{"from": 2, "to": 0}, {"from": 2, "to": 3}, {"from": 2, "to": 3}]
	}`)

	project, err := parse(data, "/proj")
	require.NoError(t, err)
	require.Equal(t, model.NodeID(3), project.Graph.Root)
	require.Len(t, project.Graph.Nodes, 4)
	require.Len(t, project.Graph.Files, 3)
	require.Equal(t, "/proj/src/main.zig", project.Entry)

	main := project.Graph.Nodes[3]
	require.Equal(t, model.Position{Line: 5, Column: 5}, main.Syntax.Start)
	require.Equal(t, model.Position{Line: 10, Column: 2}, main.Syntax.End)
	require.Equal(t, []model.NodeID{1, 4}, main.Out)

	// edges deduped
	require.Equal(t, []model.NodeID{3}, project.Graph.Nodes[4].In)

	// builtin imports skipped, local modules resolved
	require.Equal(t, map[string]string{
		"router":   "/proj/src/router.zig",
		"services": "/proj/src/services.zig",
	}, project.Modules)
}

func TestParseSkipsMissingEdges(t *testing.T) {
	data := []byte(`{
		"root": 1,
		"entry": "/proj/a.zig",
		"files": [{"path": "/proj/a.zig", "imports": [], "functions": [
			{"id": 1, "name": "main", "startLine": 0, "startCol": 0, "endLine": 2, "endCol": 1}
		]}],
		"edges": [{"from": 1, "to": 99}]
	}`)

	project, err := parse(data, "/proj")
	require.NoError(t, err)
	require.Empty(t, project.Graph.Nodes[2].Out)
}

func benchGraph() []byte {
	var helper helperGraph
	nextID := 0
	for fileIndex := 0; fileIndex < 40; fileIndex++ {
		file := helperFile{Path: fmt.Sprintf("/app/src/file%d.zig", fileIndex)}
		for index := 0; index < 60; index++ {
			file.Functions = append(file.Functions, helperFunction{
				ID:        uint64(nextID),
				Name:      fmt.Sprintf("fn%d_%d", fileIndex, index),
				StartLine: index * 10,
				EndLine:   index*10 + 8,
			})
			nextID++
		}
		helper.Files = append(helper.Files, file)
	}
	for index := 1; index < nextID; index++ {
		helper.Edges = append(helper.Edges, helperEdge{From: uint64(index - 1), To: uint64(index)})
	}
	data, err := json.Marshal(helper)
	if err != nil {
		panic(err)
	}
	return data
}

func BenchmarkParse(b *testing.B) {
	data := benchGraph()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := parse(data, "/app"); err != nil {
			b.Fatal(err)
		}
	}
}
