// Package analysis converts a Node.js project into Senbon's normalized graph.
package analysis

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/briheet/senbon/internal/model"
)

//go:embed analyze.cjs
var helperSource []byte

// helperGraph mirrors the JSON emitted by analyze.cjs.
type helperGraph struct {
	Root  uint64       `json:"root"`
	Files []helperFile `json:"files"`
	Edges []helperEdge `json:"edges"`
}

// helperEdge is one resolved call edge.
type helperEdge struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

// helperFile is one analyzed source file.
type helperFile struct {
	Path      string           `json:"path"`
	Functions []helperFunction `json:"functions"`
}

// helperFunction is one function declaration.
type helperFunction struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	StartLine int    `json:"startLine"`
	StartCol  int    `json:"startCol"`
	EndLine   int    `json:"endLine"`
	EndCol    int    `json:"endCol"`
}

// Analyze runs the embedded helper and builds the static graph for the project.
func Analyze(ctx context.Context, projectDir string) (*model.StaticGraph, error) {
	if _, err := exec.LookPath("node"); err != nil {
		return nil, fmt.Errorf("node is required for Node.js analysis: %w", err)
	}
	helperPath, err := writeHelper()
	if err != nil {
		return nil, err
	}
	defer os.Remove(helperPath)

	output, err := exec.CommandContext(ctx, "node", helperPath, projectDir).Output()
	if err != nil {
		return nil, fmt.Errorf("analysis helper failed: %w", err)
	}
	return parse(output, projectDir)
}

// parse builds a static graph from helper JSON output.
func parse(data []byte, projectDir string) (*model.StaticGraph, error) {
	var helper helperGraph
	if err := json.Unmarshal(data, &helper); err != nil {
		return nil, err
	}

	graph := &model.StaticGraph{
		Root:     model.NodeID(helper.Root),
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	graph.Packages[1] = &model.Package{Path: projectDir, Name: filepath.Base(projectDir)}

	fileIDs := make(map[string]model.FileID, len(helper.Files))
	for index, file := range helper.Files {
		id := model.FileID(index + 1)
		fileIDs[file.Path] = id
		graph.Files[id] = &model.StaticFile{ID: id, Path: file.Path, Package: 1}
		graph.Program.Files = append(graph.Program.Files, id)
	}
	for _, file := range helper.Files {
		fileID := fileIDs[file.Path]
		for _, fn := range file.Functions {
			node := &model.StaticNode{
				Name: fn.Name,
				ID:   model.NodeID(fn.ID),
				Pkg:  1,
				Syntax: model.Syntax{
					File:  fileID,
					Start: model.Position{Line: fn.StartLine + 1, Column: fn.StartCol + 1},
					End:   model.Position{Line: fn.EndLine + 1, Column: fn.EndCol + 1},
				},
			}
			graph.Nodes[node.ID] = node
			graph.Files[fileID].Functions = append(graph.Files[fileID].Functions, node.ID)
		}
	}
	for _, edge := range helper.Edges {
		from := model.NodeID(edge.From)
		to := model.NodeID(edge.To)
		if _, ok := graph.Nodes[from]; !ok {
			continue
		}
		if _, ok := graph.Nodes[to]; !ok {
			continue
		}
		graph.Nodes[from].Out = append(graph.Nodes[from].Out, to)
		graph.Nodes[to].In = append(graph.Nodes[to].In, from)
	}
	for _, node := range graph.Nodes {
		slices.Sort(node.In)
		slices.Sort(node.Out)
	}
	return graph, nil
}

// writeHelper writes the embedded analyzer to a temporary file.
func writeHelper() (string, error) {
	file, err := os.CreateTemp("", "senbon-analyze-*.cjs")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(helperSource); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}
