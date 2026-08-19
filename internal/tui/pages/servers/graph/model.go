// Package graph renders the interactive graph used by server pages.
package graph

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/harmonica"
)

const (
	framesPerSecond = 60
	frequency       = 6.0
	damping         = 1.0 // Critical damping returns nodes without oscillation.
	maxVelocity     = 24.0
	settleDistance  = 0.03
	maxImageID      = 1<<24 - 1
	dependencyScale = 0.18
	maximumScale    = 2.0
)

var imageIDs atomic.Uint32

// Kind selects the analyzed relationship represented by a graph.
type Kind uint8

const (
	// FunctionGraph shows project functions and their call relationships.
	FunctionGraph Kind = iota
	// FileGraph collapses calls into source-file relationships.
	FileGraph
	// DependencyGraph sizes function nodes by direct connectivity.
	DependencyGraph
)

type point struct {
	x float64
	y float64
}

type node struct {
	label      string
	anchor     point
	position   point
	rendered   point
	velocity   point
	id         uint64
	scale      float64
	depth      int
	row        int
	rowCount   int
	labelX     int
	labelY     int
	labelCellX int
	labelCellY int
}

// Model contains the graph layout, motion, and drag state.
type Model struct {
	renderErr      error
	dump           io.Writer
	renderer       *renderer
	owner          string
	labels         string
	nodes          []node
	edges          []edgeModel
	forceBuffer    []point
	spring         harmonica.Spring
	dragOffset     point
	dragging       int
	height         int
	width          int
	cellHeight     int
	cellWidth      int
	root           int
	originX        int
	originY        int
	layoutFrames   int
	generation     uint64
	renderSequence uint64
	revision       uint64
	frontImageID   uint32
	imageIDs       [2]uint32
	labelsDragging int
	nodeRadius     float64
	kind           Kind
	graphics       bool
	animating      bool
	visible        bool
	renderPending  bool
}

// New builds the selected graph from analyzed project code.
func New(owner string, kind Kind, source *model.RuntimeGraph, dump io.Writer) Model {
	nodes, edges, root := build(source, kind)
	m := Model{
		owner:          owner,
		kind:           kind,
		dump:           dump,
		nodes:          nodes,
		edges:          edges,
		root:           root,
		spring:         harmonica.NewSpring(harmonica.FPS(framesPerSecond), frequency, damping),
		cellWidth:      fallbackCellWidth,
		cellHeight:     fallbackCellHeight,
		nodeRadius:     fallbackNodeRadius,
		graphics:       supportsGraphics(),
		imageIDs:       [2]uint32{nextImageID(), nextImageID()},
		dragging:       -1,
		labelsDragging: -2,
	}
	m.renderer = newRenderer(owner, edges, dump)
	m.trace("initialized graphics=%t nodes=%d edges=%d term=%q term_program=%q kitty_window=%t",
		m.graphics, len(nodes), len(edges), os.Getenv("TERM"), os.Getenv("TERM_PROGRAM"), os.Getenv("KITTY_WINDOW_ID") != "")
	return m
}

// Init waits for the viewport before rendering or animating.
func (Model) Init() tea.Cmd { return nil }

// Revision changes when the graph's native terminal layer changes.
func (m Model) Revision() uint64 { return m.revision }

func build(source *model.RuntimeGraph, kind Kind) ([]node, []edgeModel, int) {
	if kind == FileGraph {
		return buildFiles(source)
	}
	return buildFunctions(source, kind == DependencyGraph)
}

func buildFunctions(source *model.RuntimeGraph, dependencies bool) ([]node, []edgeModel, int) {
	if source == nil || source.Static == nil {
		return nil, nil, -1
	}

	localIDs := make([]model.NodeID, 0, len(source.Nodes))
	for id, runtimeNode := range source.Nodes {
		if runtimeNode != nil && runtimeNode.Static != nil {
			localIDs = append(localIDs, id)
		}
	}
	slices.Sort(localIDs)
	if len(localIDs) == 0 {
		return nil, nil, -1
	}

	ids := append([]model.NodeID(nil), localIDs...)
	if dependencies {
		seen := make(map[model.NodeID]struct{}, len(localIDs))
		for _, id := range localIDs {
			seen[id] = struct{}{}
		}
		for _, id := range localIDs {
			function := source.Nodes[id].Static
			for _, targets := range [][]model.NodeID{function.Out, function.Function.References, function.Function.AnonFuncs} {
				for _, target := range targets {
					if _, exists := seen[target]; exists || source.Static.Nodes[target] == nil {
						continue
					}
					seen[target] = struct{}{}
					ids = append(ids, target)
				}
			}
		}
		slices.Sort(ids)
	}

	indices := make(map[model.NodeID]int, len(ids))
	names := make(map[string]int, len(ids))
	for index, id := range ids {
		indices[id] = index
		names[source.Static.Nodes[id].Name]++
	}
	nodes := make([]node, len(ids))
	for index, id := range ids {
		static := source.Static.Nodes[id]
		label := functionLabel(source.Static, static, names[static.Name] > 1)
		if _, local := source.Nodes[id]; dependencies && !local {
			label = dependencyLabel(source.Static, static, names[static.Name] > 1)
		}
		nodes[index] = node{id: uint64(id), label: label, scale: 1}
	}

	edgeCapacity := 0
	for _, id := range localIDs {
		static := source.Nodes[id].Static
		edgeCapacity += len(static.Out) + len(static.Function.References) + len(static.Function.AnonFuncs)
	}
	edges := make([]edgeModel, 0, edgeCapacity)
	for _, id := range localIDs {
		from := indices[id]
		static := source.Nodes[id].Static
		for _, targets := range [][]model.NodeID{static.Out, static.Function.References, static.Function.AnonFuncs} {
			for _, target := range targets {
				if to, ok := indices[target]; ok {
					edges = append(edges, edgeModel{from: from, to: to})
				}
			}
		}
	}
	edges = normalizeEdges(edges)
	if dependencies {
		scaleDependencies(nodes, edges)
	}
	root := graphRoot(source, localIDs, indices)
	assignDepths(nodes, edges, root)
	return nodes, edges, root
}

func buildFiles(source *model.RuntimeGraph) ([]node, []edgeModel, int) {
	if source == nil || source.Static == nil {
		return nil, nil, -1
	}

	ids := make([]model.FileID, 0, len(source.Files))
	for id, runtimeFile := range source.Files {
		if runtimeFile != nil && runtimeFile.Static != nil {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	if len(ids) == 0 {
		return nil, nil, -1
	}

	indices := make(map[model.FileID]int, len(ids))
	names := make(map[string]int, len(ids))
	for index, id := range ids {
		indices[id] = index
		names[filepath.Base(source.Files[id].Static.Path)]++
	}
	nodes := make([]node, len(ids))
	edges := make([]edgeModel, 0, len(ids))
	for from, id := range ids {
		file := source.Files[id].Static
		nodes[from] = node{id: uint64(id), label: fileLabel(file.Path, names[filepath.Base(file.Path)] > 1), scale: 1}
		for _, functionID := range file.Functions {
			function := source.Nodes[functionID]
			if function == nil || function.Static == nil {
				continue
			}
			for _, targets := range [][]model.NodeID{function.Static.Out, function.Static.Function.References, function.Static.Function.AnonFuncs} {
				for _, target := range targets {
					dependency := source.Static.Nodes[target]
					if dependency == nil {
						continue
					}
					if to, ok := indices[dependency.Syntax.File]; ok && to != from {
						edges = append(edges, edgeModel{from: from, to: to})
					}
				}
			}
		}
	}
	edges = normalizeEdges(edges)
	root := fileRoot(source, ids, indices, edges)
	assignDepths(nodes, edges, root)
	return nodes, edges, root
}

func normalizeEdges(edges []edgeModel) []edgeModel {
	slices.SortFunc(edges, func(left, right edgeModel) int {
		if left.from != right.from {
			return left.from - right.from
		}
		return left.to - right.to
	})
	return slices.CompactFunc(edges, func(left, right edgeModel) bool {
		return left.from == right.from && left.to == right.to
	})
}

func scaleDependencies(nodes []node, edges []edgeModel) {
	degrees := make([]int, len(nodes))
	for _, edge := range edges {
		if edge.from != edge.to {
			degrees[edge.from]++
			degrees[edge.to]++
		}
	}
	for index := range nodes {
		nodes[index].scale = min(maximumScale, 1+math.Sqrt(float64(degrees[index]))*dependencyScale)
	}
}

func functionLabel(graph *model.StaticGraph, function *model.StaticNode, duplicate bool) string {
	name := function.Name
	file := graph.Files[function.Syntax.File]
	if name == "" {
		name = "function-" + fmt.Sprint(function.ID)
	}
	if !duplicate || file == nil {
		return name
	}
	return fmt.Sprintf("%s @ %s:%d", name, filepath.Base(file.Path), function.Syntax.Start.Line)
}

func dependencyLabel(graph *model.StaticGraph, function *model.StaticNode, duplicate bool) string {
	name := functionLabel(graph, function, duplicate)
	pkg := graph.Packages[function.Pkg]
	if pkg == nil {
		return name
	}
	prefix := pkg.Name
	if prefix == "" {
		prefix = filepath.Base(pkg.Path)
	}
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func fileLabel(path string, duplicate bool) string {
	name := filepath.Base(path)
	if !duplicate {
		return name
	}
	return filepath.Join(filepath.Base(filepath.Dir(path)), name)
}

func graphRoot(source *model.RuntimeGraph, ids []model.NodeID, indices map[model.NodeID]int) int {
	if root, ok := indices[source.Static.Root]; ok {
		return root
	}
	localRoot := func(id model.NodeID) bool {
		for _, caller := range source.Nodes[id].Static.In {
			if _, ok := indices[caller]; ok {
				return false
			}
		}
		return true
	}
	for _, id := range ids {
		if source.Nodes[id].Static.Name == "main" && localRoot(id) {
			return indices[id]
		}
	}
	for _, id := range ids {
		if localRoot(id) {
			return indices[id]
		}
	}
	return 0
}

func fileRoot(source *model.RuntimeGraph, ids []model.FileID, indices map[model.FileID]int, edges []edgeModel) int {
	if root := source.Static.Nodes[source.Static.Root]; root != nil {
		if index, ok := indices[root.Syntax.File]; ok {
			return index
		}
	}
	for _, id := range ids {
		for _, functionID := range source.Files[id].Static.Functions {
			if function := source.Nodes[functionID]; function != nil && function.Static.Name == "main" {
				return indices[id]
			}
		}
	}
	incoming := make([]bool, len(ids))
	for _, edge := range edges {
		incoming[edge.to] = true
	}
	for index := range ids {
		if !incoming[index] {
			return index
		}
	}
	return 0
}

func assignDepths(nodes []node, edges []edgeModel, root int) {
	// Edges are sorted by source, so offsets provide adjacency without slices per node.
	offsets := make([]int, len(nodes)+1)
	for _, edge := range edges {
		offsets[edge.from+1]++
	}
	for index := 1; index < len(offsets); index++ {
		offsets[index] += offsets[index-1]
	}
	for index := range nodes {
		nodes[index].depth = -1
	}
	queue := make([]int, 0, len(nodes))
	walk := func(start, depth int) {
		queue = append(queue[:0], start)
		nodes[start].depth = depth
		for head := 0; head < len(queue); head++ {
			from := queue[head]
			for _, edge := range edges[offsets[from]:offsets[from+1]] {
				if nodes[edge.to].depth < 0 {
					nodes[edge.to].depth = nodes[from].depth + 1
					queue = append(queue, edge.to)
				}
			}
		}
	}
	walk(root, 0)
	for index := range nodes {
		if nodes[index].depth < 0 {
			walk(index, 1)
		}
	}

	rows := make([]int, 2*(len(nodes)+1))
	counts := rows[:len(nodes)+1]
	rows = rows[len(nodes)+1:]
	for _, node := range nodes {
		counts[node.depth]++
	}
	for index := range nodes {
		depth := nodes[index].depth
		nodes[index].row = rows[depth]
		nodes[index].rowCount = counts[depth]
		rows[depth]++
	}
}

func (m Model) trace(format string, values ...any) {
	if m.dump != nil {
		_, _ = fmt.Fprintf(m.dump, "graph[%s] "+format+"\n", append([]any{m.owner}, values...)...)
	}
}

func distance(first, second point) float64 {
	return math.Hypot(second.x-first.x, second.y-first.y)
}

func nextImageID() uint32 {
	id := imageIDs.Add(1)
	if id > maxImageID {
		id = (id-1)%maxImageID + 1
	}
	return id
}
