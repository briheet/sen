package graph

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/basicfont"
)

func TestGraphBuildsProjectFunctionsAndCalls(t *testing.T) {
	graph := New("api", testRuntimeGraph(), nil)

	require.Equal(t, []model.NodeID{1, 2, 3}, []model.NodeID{
		graph.nodes[0].id,
		graph.nodes[1].id,
		graph.nodes[2].id,
	})
	require.Equal(t, []string{"main", "routes", "handler"}, []string{
		graph.nodes[0].label,
		graph.nodes[1].label,
		graph.nodes[2].label,
	})
	require.Equal(t, 0, graph.root)
	require.Equal(t, []edgeModel{{from: 0, to: 1}, {from: 0, to: 2}, {from: 1, to: 2}}, graph.edges)
}

func TestGraphDisambiguatesDuplicateFunctionNames(t *testing.T) {
	source := testRuntimeGraph()
	source.Static.Nodes[2].Name = "handler"
	source.Nodes[2].Static.Name = "handler"
	graph := New("api", source, nil)

	require.Equal(t, "handler @ main.go:8", graph.nodes[1].label)
	require.Equal(t, "handler @ handlers.go:4", graph.nodes[2].label)
}

func TestGraphRendersKittyPlaceholders(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, command := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	graph.ready = true
	view := graph.View()

	require.NotNil(t, command)
	require.Equal(t, 80, lipgloss.Width(view))
	require.Equal(t, 16, lipgloss.Height(view))
	require.Contains(t, view, string(kitty.Placeholder))
	require.NotContains(t, view, "router")
}

func TestGraphTransmitsPNGBeforeVirtualPlacement(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	message := graph.upload()()
	raw, ok := message.(tea.RawMsg)
	require.True(t, ok)
	sequence := fmt.Sprint(raw.Msg)

	for _, option := range []string{"f=100", "q=2", "a=p", "c=40", "r=10", "U=1", "C=1"} {
		require.Contains(t, sequence, option)
	}
	require.Less(t, strings.Index(sequence, "f=100"), strings.Index(sequence, "a=p"))
}

func TestGraphDebugLogKeepsTerminalErrors(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	var dump bytes.Buffer
	graph, _ := New("api", testRuntimeGraph(), &dump).Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	raw := graph.upload()().(tea.RawMsg)

	require.Contains(t, fmt.Sprint(raw.Msg), "q=1")
	require.Contains(t, dump.String(), "nodes=3 edges=3")
	require.Contains(t, dump.String(), "upload queued")
}

func TestGraphDrawsEdgesBeforeNodes(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	canvas := graph.renderImage()
	edge := graph.edges[0]
	from, to := edgePoints(graph.nodes[edge.from], graph.nodes[edge.to])
	middle := point{x: (from.x + to.x) / 2, y: (from.y + to.y) / 2}

	edgePixels := 0
	for y := int(middle.y) - 2; y <= int(middle.y)+2; y++ {
		for x := int(middle.x) - 2; x <= int(middle.x)+2; x++ {
			if canvas.RGBAAt(x, y).A != 0 {
				edgePixels++
			}
		}
	}
	require.Positive(t, edgePixels)
	require.NotZero(t, canvas.RGBAAt(int(from.x), int(from.y)).A)

	label := labelPoint(canvas.Bounds(), pixelPoint(graph.nodes[0].position), graph.nodes[0].label)
	labelStart := label.X.Ceil()
	labelPixels := 0
	for y := max(0, label.Y.Ceil()-basicfont.Face7x13.Height); y < min(canvas.Bounds().Dy(), label.Y.Ceil()+1); y++ {
		for x := labelStart; x < min(canvas.Bounds().Dx(), labelStart+glyphWidth*len(graph.nodes[0].label)); x++ {
			if canvas.RGBAAt(x, y).A != 0 {
				labelPixels++
			}
		}
	}
	require.Positive(t, labelPixels)
}

func TestGraphSettlesAndPinsRoot(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	graph = settleLayout(graph)

	require.False(t, graph.layoutSettling)
	require.False(t, graph.animating)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
	for _, node := range graph.nodes {
		require.Equal(t, node.anchor, node.position)
		require.GreaterOrEqual(t, node.position.x, 0.0)
		require.GreaterOrEqual(t, node.position.y, 0.0)
	}
}

func TestRootCannotBeDragged(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	graph = settleLayout(graph)
	root := screenPoint(graph.nodes[graph.root].position)

	graph, command := graph.Update(tea.MouseClickMsg{X: int(root.x), Y: int(root.y), Button: tea.MouseLeft})

	require.Nil(t, command)
	require.Equal(t, -1, graph.dragging)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
}

func TestDraggedNodeJigglesNeighborAndReturnsToAnchor(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	graph = settleLayout(graph)
	dragged := 1
	neighbor := 2
	anchor := graph.nodes[dragged].anchor
	neighborBefore := graph.nodes[neighbor].position
	target := screenPoint(graph.nodes[neighbor].position)
	draggedPosition := screenPoint(graph.nodes[dragged].position)
	click := tea.MouseClickMsg{
		X:      int(math.Round(draggedPosition.x)),
		Y:      int(math.Round(draggedPosition.y)),
		Button: tea.MouseLeft,
	}

	graph, command := graph.Update(click)
	require.NotNil(t, command)
	require.Equal(t, dragged, graph.dragging)
	graph, _ = graph.Update(tea.MouseMotionMsg{X: int(target.x), Y: int(target.y), Button: tea.MouseLeft})
	for range 5 {
		graph, _ = graph.Update(frameMsg{owner: "api", generation: graph.generation})
	}
	require.NotEqual(t, neighborBefore, graph.nodes[neighbor].position)

	graph, _ = graph.Update(tea.MouseReleaseMsg{X: int(target.x), Y: int(target.y), Button: tea.MouseLeft})
	for frame := 0; frame < maximumReturnFrames+1 && graph.animating; frame++ {
		graph, _ = graph.Update(frameMsg{owner: "api", generation: graph.generation})
	}
	require.False(t, graph.animating)
	require.Equal(t, anchor, graph.nodes[dragged].position)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
}

func TestResizeRestartsLayout(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	graph = settleLayout(graph)
	generation := graph.generation

	graph, command := graph.Update(tea.WindowSizeMsg{Width: 100, Height: 24})

	require.NotNil(t, command)
	require.True(t, graph.layoutSettling)
	require.Greater(t, graph.generation, generation)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
}

func TestGraphUsesUniqueImageIDs(t *testing.T) {
	first := New("api", testRuntimeGraph(), nil)
	second := New("worker", testRuntimeGraph(), nil)
	require.NotZero(t, first.imageID)
	require.NotEqual(t, first.imageID, second.imageID)
}

func TestUnsupportedTerminalHasTextFallback(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "Alacritty")
	t.Setenv("KITTY_WINDOW_ID", "")
	graph, command := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 60, Height: 12})
	view := graph.View()

	require.Nil(t, command)
	require.Contains(t, view, "requires Ghostty or Kitty")
	require.False(t, strings.ContainsRune(view, kitty.Placeholder))
}

func TestEmptyGraphHasTextFallback(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, command := New("api", nil, nil).Update(tea.WindowSizeMsg{Width: 60, Height: 12})

	require.Nil(t, command)
	require.Contains(t, graph.View(), "No project functions found")
}

func TestGraphIgnoresFramesForAnotherServer(t *testing.T) {
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	before := graph.nodes[1].position

	graph, command := graph.Update(frameMsg{owner: "worker", generation: graph.generation})

	require.Nil(t, command)
	require.Equal(t, before, graph.nodes[1].position)
}

func BenchmarkGraphUpload(b *testing.B) {
	b.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	raw := graph.upload()().(tea.RawMsg)
	wireBytes := len(fmt.Sprint(raw.Msg))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = graph.upload()
	}
	b.ReportMetric(float64(wireBytes), "wire-B/frame")
}

func BenchmarkGraphBuild(b *testing.B) {
	source := testRuntimeGraph()
	b.ReportAllocs()
	for b.Loop() {
		_ = New("api", source, nil)
	}
}

func BenchmarkGraphView(b *testing.B) {
	b.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", testRuntimeGraph(), nil).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	graph.ready = true
	b.ReportAllocs()
	for b.Loop() {
		if graph.View() == "" {
			b.Fatal("empty graph view")
		}
	}
}

func settleLayout(graph Model) Model {
	for frame := 0; frame < maximumLayoutFrames+1 && graph.animating; frame++ {
		graph, _ = graph.Update(frameMsg{owner: graph.owner, generation: graph.generation})
	}
	return graph
}

func testRuntimeGraph() *model.RuntimeGraph {
	static := &model.StaticGraph{
		Root: 100,
		Nodes: map[model.NodeID]*model.StaticNode{
			1: {
				ID:     1,
				Name:   "main",
				Syntax: model.Syntax{File: 1, Start: model.Position{Line: 3}},
				In:     []model.NodeID{100},
				Out:    []model.NodeID{99, 2},
				Function: model.FunctionBody{
					AnonFuncs: []model.NodeID{3},
				},
			},
			2: {
				ID:     2,
				Name:   "routes",
				Syntax: model.Syntax{File: 1, Start: model.Position{Line: 8}},
				In:     []model.NodeID{1},
				Function: model.FunctionBody{
					References: []model.NodeID{3},
				},
			},
			3: {
				ID:     3,
				Name:   "handler",
				Syntax: model.Syntax{File: 2, Start: model.Position{Line: 4}},
				In:     []model.NodeID{1, 2},
			},
			99:  {ID: 99, Name: "external"},
			100: {ID: 100, Name: "root", Out: []model.NodeID{1}},
		},
		Files: map[model.FileID]*model.StaticFile{
			1: {ID: 1, Path: "/project/main.go", Functions: []model.NodeID{1, 2}},
			2: {ID: 2, Path: "/project/handlers.go", Functions: []model.NodeID{3}},
		},
	}
	return &model.RuntimeGraph{
		Static: static,
		Nodes: map[model.NodeID]*model.Node{
			1: {Static: static.Nodes[1]},
			2: {Static: static.Nodes[2]},
			3: {Static: static.Nodes[3]},
		},
	}
}
