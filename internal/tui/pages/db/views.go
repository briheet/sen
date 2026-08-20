package db

import (
	"maps"
	"path/filepath"
	"strings"

	"github.com/briheet/sen/internal/model"
)

// Database adapters expose stable entity kinds through synthetic path
// segments. Both views preserve the engine's IDs, so one telemetry snapshot
// can update them without duplicating collection or attribution.
func databaseViews(source *model.RuntimeGraph) (*model.RuntimeGraph, *model.RuntimeGraph) {
	return databaseView(source, "stmt"), databaseView(source, "table")
}

func databaseView(source *model.RuntimeGraph, entity string) *model.RuntimeGraph {
	if source == nil || source.Static == nil {
		return nil
	}
	static := source.Static
	root := static.Nodes[static.Root]
	if root == nil {
		return nil
	}

	includedFiles := map[model.FileID]bool{root.Syntax.File: true}
	includedNodes := map[model.NodeID]bool{static.Root: true}
	segment := "/" + entity + "/"
	for id, file := range static.Files {
		if file == nil || !strings.Contains(filepath.ToSlash(file.Path), segment) {
			continue
		}
		includedFiles[id] = true
		for _, node := range file.Functions {
			includedNodes[node] = true
		}
	}

	view := &model.StaticGraph{
		Root: static.Root, Nodes: make(map[model.NodeID]*model.StaticNode, len(includedNodes)),
		Files:    make(map[model.FileID]*model.StaticFile, len(includedFiles)),
		Packages: maps.Clone(static.Packages),
	}
	for id := range includedNodes {
		node := static.Nodes[id]
		if node == nil {
			continue
		}
		clone := *node
		clone.In = included(node.In, includedNodes)
		clone.Out = included(node.Out, includedNodes)
		clone.Function.AnonFuncs = included(node.Function.AnonFuncs, includedNodes)
		clone.Function.References = included(node.Function.References, includedNodes)
		if clone.Parent != nil && !includedNodes[*clone.Parent] {
			clone.Parent = nil
		}
		view.Nodes[id] = &clone
	}
	for id := range includedFiles {
		file := static.Files[id]
		if file == nil {
			continue
		}
		clone := *file
		clone.Functions = included(file.Functions, includedNodes)
		clone.Calls = included(file.Calls, includedFiles)
		clone.CalledBy = included(file.CalledBy, includedFiles)
		view.Files[id] = &clone
	}

	modulePath := static.Packages[root.Pkg].Path
	return model.BuildRuntimeGraph(modulePath, view)
}

func included[T comparable](values []T, set map[T]bool) []T {
	result := make([]T, 0, len(values))
	for _, value := range values {
		if set[value] {
			result = append(result, value)
		}
	}
	return result
}
