// Package graph renders the interactive graph used by server pages.
package graph

import (
	"fmt"
	"image"
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
	framesPerSecond = 30
	frequency       = 6.0
	damping         = 0.25
	maxVelocity     = 24.0
	settleDistance  = 0.03
	maxImageID      = 1<<24 - 1
)

var imageIDs atomic.Uint32

type point struct {
	x float64
	y float64
}

type node struct {
	label    string
	seed     point
	anchor   point
	position point
	velocity point
	id       model.NodeID
	depth    int
	row      int
	rowCount int
}

// Model contains the graph layout, motion, and drag state.
type Model struct {
	renderErr      error
	dump           io.Writer
	canvas         *image.RGBA
	owner          string
	placeholder    string
	nodes          []node
	edges          []edgeModel
	forceBuffer    []point
	spring         harmonica.Spring
	dragOffset     point
	dragging       int
	height         int
	width          int
	root           int
	layoutFrames   int
	generation     uint64
	imageID        uint32
	graphics       bool
	ready          bool
	animating      bool
	layoutSettling bool
}

// New builds a function graph from analyzed project code.
func New(owner string, source *model.RuntimeGraph, dump io.Writer) Model {
	nodes, edges, root := build(source)
	m := Model{
		owner:    owner,
		dump:     dump,
		nodes:    nodes,
		edges:    edges,
		root:     root,
		spring:   harmonica.NewSpring(harmonica.FPS(framesPerSecond), frequency, damping),
		graphics: supportsGraphics(),
		imageID:  nextImageID(),
		dragging: -1,
	}
	m.trace("initialized graphics=%t nodes=%d edges=%d term=%q term_program=%q kitty_window=%t",
		m.graphics, len(nodes), len(edges), os.Getenv("TERM"), os.Getenv("TERM_PROGRAM"), os.Getenv("KITTY_WINDOW_ID") != "")
	return m
}

// Init waits for the viewport before rendering or animating.
func (Model) Init() tea.Cmd { return nil }

func build(source *model.RuntimeGraph) ([]node, []edgeModel, int) {
	if source == nil || source.Static == nil {
		return nil, nil, -1
	}

	ids := make([]model.NodeID, 0, len(source.Nodes))
	for id, runtimeNode := range source.Nodes {
		if runtimeNode != nil && runtimeNode.Static != nil {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	if len(ids) == 0 {
		return nil, nil, -1
	}

	indices := make(map[model.NodeID]int, len(ids))
	names := make(map[string]int, len(ids))
	for index, id := range ids {
		indices[id] = index
		names[source.Nodes[id].Static.Name]++
	}
	nodes := make([]node, len(ids))
	for index, id := range ids {
		static := source.Nodes[id].Static
		nodes[index] = node{id: id, label: functionLabel(source.Static, static, names[static.Name] > 1)}
	}

	edgeCapacity := 0
	for _, id := range ids {
		static := source.Nodes[id].Static
		edgeCapacity += len(static.Out) + len(static.Function.References) + len(static.Function.AnonFuncs)
	}
	edges := make([]edgeModel, 0, edgeCapacity)
	for from, id := range ids {
		static := source.Nodes[id].Static
		for _, targets := range [][]model.NodeID{static.Out, static.Function.References, static.Function.AnonFuncs} {
			for _, target := range targets {
				if to, ok := indices[target]; ok {
					edges = append(edges, edgeModel{from: from, to: to})
				}
			}
		}
	}
	slices.SortFunc(edges, func(left, right edgeModel) int {
		if left.from != right.from {
			return left.from - right.from
		}
		return left.to - right.to
	})
	edges = slices.CompactFunc(edges, func(left, right edgeModel) bool {
		return left.from == right.from && left.to == right.to
	})
	root := graphRoot(source, ids, indices)
	assignDepths(nodes, edges, root)
	return nodes, edges, root
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
