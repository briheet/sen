package model

import (
	"path/filepath"
	"strings"
	"sync"

	graph "github.com/briheet/senbon/internal/analysis"
	runtimepprof "github.com/briheet/senbon/internal/runtime/pprof"
	runtimetrace "github.com/briheet/senbon/internal/runtime/trace"
)

type sourceFrame struct {
	file string
	line int64
}

const (
	maxPooledTargets     = 4096
	packagePathSeparator = "/"
)

type targetWorkspace struct {
	frames    []sourceFrame
	nodes     []graph.NodeID
	files     []graph.FileID
	seenNodes map[graph.NodeID]struct{}
	seenFiles map[graph.FileID]struct{}
}

var targetWorkspaces = sync.Pool{New: func() any {
	return &targetWorkspace{
		seenNodes: make(map[graph.NodeID]struct{}),
		seenFiles: make(map[graph.FileID]struct{}),
	}
}}

type frameMapper struct {
	static   *graph.Graph
	paths    map[string]graph.FileID
	relative map[graph.FileID]string
}

func newMapper(modulePath string, static *graph.Graph, result *RuntimeGraph) *frameMapper {
	mapper := &frameMapper{
		static:   static,
		paths:    make(map[string]graph.FileID),
		relative: make(map[graph.FileID]string),
	}
	for id, file := range static.Files {
		pkg := static.Packages[file.Package]
		if pkg.Path != modulePath && !strings.HasPrefix(pkg.Path, modulePath+packagePathSeparator) {
			continue
		}
		path := filepath.Clean(file.Path)
		relative := strings.TrimPrefix(strings.TrimPrefix(pkg.Path, modulePath), packagePathSeparator)
		if relative != "" {
			relative += packagePathSeparator
		}
		relative += filepath.Base(path)
		mapper.paths[path] = id
		mapper.relative[id] = relative
		result.Files[id] = &File{Static: file, Metrics: make(CodeMetrics)}
	}
	for id, node := range static.Nodes {
		if _, ok := result.Files[node.Syntax.File]; ok {
			result.Nodes[id] = &Node{Static: node, Metrics: make(CodeMetrics)}
		}
	}
	return mapper
}

func (m *frameMapper) profileTargets(profile *runtimepprof.Profile, sample runtimepprof.Sample, workspace *targetWorkspace) ([]graph.NodeID, []graph.FileID) {
	workspace.reset()
	for _, id := range sample.Stack {
		location, ok := profile.Locations[id]
		if !ok {
			continue
		}
		for _, frame := range location.Frames {
			workspace.frames = append(workspace.frames, sourceFrame{file: frame.File, line: frame.Line})
		}
	}
	return m.targets(workspace)
}

func (m *frameMapper) traceTargets(trace *runtimetrace.Trace, stackID runtimetrace.StackID, workspace *targetWorkspace) ([]graph.NodeID, []graph.FileID) {
	workspace.reset()
	stack, ok := trace.Stacks[stackID]
	if !ok {
		return nil, nil
	}
	for _, frame := range stack.Frames {
		workspace.frames = append(workspace.frames, sourceFrame{file: frame.File, line: int64(frame.Line)})
	}
	return m.targets(workspace)
}

func (m *frameMapper) targets(workspace *targetWorkspace) ([]graph.NodeID, []graph.FileID) {
	for _, frame := range workspace.frames {
		fileID, ok := m.file(frame.file)
		if !ok {
			continue
		}
		if _, ok := workspace.seenFiles[fileID]; !ok {
			workspace.seenFiles[fileID] = struct{}{}
			workspace.files = append(workspace.files, fileID)
		}
		if nodeID, ok := m.node(fileID, frame.line); ok {
			if _, seen := workspace.seenNodes[nodeID]; !seen {
				workspace.seenNodes[nodeID] = struct{}{}
				workspace.nodes = append(workspace.nodes, nodeID)
			}
		}
	}
	return workspace.nodes, workspace.files
}

func acquireTargetWorkspace() *targetWorkspace {
	return targetWorkspaces.Get().(*targetWorkspace)
}

func releaseTargetWorkspace(workspace *targetWorkspace) {
	workspace.reset()
	if cap(workspace.frames) > maxPooledTargets {
		workspace.frames = nil
	}
	if cap(workspace.nodes) > maxPooledTargets {
		workspace.nodes = nil
	}
	if cap(workspace.files) > maxPooledTargets {
		workspace.files = nil
	}
	targetWorkspaces.Put(workspace)
}

func (workspace *targetWorkspace) reset() {
	workspace.frames = workspace.frames[:0]
	workspace.nodes = workspace.nodes[:0]
	workspace.files = workspace.files[:0]
	if len(workspace.seenNodes) > maxPooledTargets {
		workspace.seenNodes = make(map[graph.NodeID]struct{})
	} else {
		clear(workspace.seenNodes)
	}
	if len(workspace.seenFiles) > maxPooledTargets {
		workspace.seenFiles = make(map[graph.FileID]struct{})
	} else {
		clear(workspace.seenFiles)
	}
}

func (m *frameMapper) file(path string) (graph.FileID, bool) {
	if path == "" {
		return 0, false
	}
	clean := filepath.Clean(path)
	if id, ok := m.paths[clean]; ok {
		return id, true
	}

	candidate := filepath.ToSlash(filepath.Clean(path))
	var match graph.FileID
	longest := 0
	for id, relative := range m.relative {
		if candidate == relative || strings.HasSuffix(candidate, "/"+relative) {
			if len(relative) > longest {
				match = id
				longest = len(relative)
			}
		}
	}
	return match, match != 0
}

func (m *frameMapper) node(fileID graph.FileID, line int64) (graph.NodeID, bool) {
	file := m.static.Files[fileID]
	if file == nil || line <= 0 {
		return 0, false
	}
	var match graph.NodeID
	bestSpan := int(^uint(0) >> 1)
	bestColumn := -1
	for _, id := range file.Functions {
		node, ok := m.static.Nodes[id]
		if !ok || node.Syntax.Start.Line <= 0 || node.Syntax.End.Line <= 0 {
			continue
		}
		if line < int64(node.Syntax.Start.Line) || line > int64(node.Syntax.End.Line) {
			continue
		}
		span := node.Syntax.End.Line - node.Syntax.Start.Line
		column := node.Syntax.Start.Column
		if span < bestSpan || span == bestSpan && column > bestColumn || span == bestSpan && column == bestColumn && (match == 0 || id < match) {
			match, bestSpan, bestColumn = id, span, column
		}
	}
	return match, match != 0
}
