package analysis

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	data := []byte(`{
		"root": 0,
		"files": [
			{"path": "/proj/index.js", "functions": [
				{"id": 0, "name": "index.js", "startLine": 0, "startCol": 0, "endLine": 9, "endCol": 0},
				{"id": 1, "name": "main", "startLine": 1, "startCol": 0, "endLine": 5, "endCol": 0},
				{"id": 2, "name": "work", "startLine": 6, "startCol": 0, "endLine": 8, "endCol": 0}
			]},
			{"path": "/proj/lib.js", "functions": [
				{"id": 3, "name": "helper", "startLine": 0, "startCol": 0, "endLine": 2, "endCol": 0}
			]}
		],
		"edges": [{"from": 0, "to": 1}, {"from": 1, "to": 2}, {"from": 2, "to": 3}]
	}`)

	graph, err := parse(data, "/proj")
	require.NoError(t, err)
	require.Equal(t, model.NodeID(0), graph.Root)
	require.Len(t, graph.Nodes, 4)
	require.Len(t, graph.Files, 2)
	require.Equal(t, "/proj", graph.Packages[1].Path)

	root := graph.Nodes[0]
	require.Equal(t, "index.js", root.Name)
	require.Equal(t, model.FileID(1), root.Syntax.File)
	require.Equal(t, []model.NodeID{1}, root.Out)

	main := graph.Nodes[1]
	require.Equal(t, model.Position{Line: 2, Column: 1}, main.Syntax.Start)
	require.Equal(t, model.Position{Line: 6, Column: 1}, main.Syntax.End)
	require.Equal(t, []model.NodeID{0}, main.In)
	require.Equal(t, []model.NodeID{2}, main.Out)

	helper := graph.Nodes[3]
	require.Equal(t, []model.NodeID{2}, helper.In)
	require.Len(t, graph.Files[2].Functions, 1)
	require.Equal(t, []model.NodeID{0, 1, 2}, graph.Files[1].Functions)
}

func TestParseSkipsMissingEdges(t *testing.T) {
	data := []byte(`{
		"root": 0,
		"files": [{"path": "/proj/a.js", "functions": [{"id": 0, "name": "a.js", "startLine": 0, "startCol": 0, "endLine": 1, "endCol": 0}]}],
		"edges": [{"from": 0, "to": 99}, {"from": 99, "to": 0}]
	}`)

	graph, err := parse(data, "/proj")
	require.NoError(t, err)
	require.Empty(t, graph.Nodes[0].Out)
	require.Empty(t, graph.Nodes[0].In)
}

// benchGraph builds helper output for 40 files with 60 functions each.
func benchGraph() []byte {
	var helper helperGraph
	nextID := 0
	for fileIndex := 0; fileIndex < 40; fileIndex++ {
		file := helperFile{Path: fmt.Sprintf("/app/src/file%d.js", fileIndex)}
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
