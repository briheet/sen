package tigerbeetle

import (
	"maps"
	"path/filepath"
	"strings"

	"github.com/briheet/sen/internal/model"
)

// The adapter exposes both entity kinds in one attribution graph. Each page
// view retains the original IDs so one engine snapshot updates both graphs.
func tigerBeetleViews(source *model.RuntimeGraph) (*model.RuntimeGraph, *model.RuntimeGraph) {
	return tigerBeetleView(source, "operation"), tigerBeetleView(source, "replica")
}

func tigerBeetleView(source *model.RuntimeGraph, entity string) *model.RuntimeGraph {
	if source == nil || source.Static == nil {
		return nil
	}
	static := source.Static
	root := static.Nodes[static.Root]
	if root == nil {
		return nil
	}
	files := map[model.FileID]bool{root.Syntax.File: true}
	nodes := map[model.NodeID]bool{static.Root: true}
	segment := "/" + entity + "/"
	for id, file := range static.Files {
		if file == nil || !strings.Contains(filepath.ToSlash(file.Path), segment) {
			continue
		}
		files[id] = true
		for _, node := range file.Functions {
			nodes[node] = true
		}
	}
	view := &model.StaticGraph{
		Root: static.Root, Nodes: make(map[model.NodeID]*model.StaticNode, len(nodes)),
		Files: make(map[model.FileID]*model.StaticFile, len(files)), Packages: maps.Clone(static.Packages),
	}
	for id := range nodes {
		node := static.Nodes[id]
		if node == nil {
			continue
		}
		clone := *node
		clone.In = selected(node.In, nodes)
		clone.Out = selected(node.Out, nodes)
		clone.Function.AnonFuncs = selected(node.Function.AnonFuncs, nodes)
		clone.Function.References = selected(node.Function.References, nodes)
		view.Nodes[id] = &clone
	}
	for id := range files {
		file := static.Files[id]
		if file == nil {
			continue
		}
		clone := *file
		clone.Functions = selected(file.Functions, nodes)
		clone.Calls = selected(file.Calls, files)
		clone.CalledBy = selected(file.CalledBy, files)
		view.Files[id] = &clone
	}
	return model.BuildRuntimeGraph(static.Packages[root.Pkg].Path, view)
}

func selected[T comparable](values []T, set map[T]bool) []T {
	result := make([]T, 0, len(values))
	for _, value := range values {
		if set[value] {
			result = append(result, value)
		}
	}
	return result
}
