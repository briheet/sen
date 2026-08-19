package graph

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/stretchr/testify/require"
)

func TestGraphBuildsProjectFunctionsAndCalls(t *testing.T) {
	graph := New("api", FunctionGraph, testRuntimeGraph(), nil)

	require.Equal(t, []uint64{1, 2, 3}, []uint64{
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

func TestGraphBuildsProjectFilesAndCalls(t *testing.T) {
	graph := New("api", FileGraph, testRuntimeGraph(), nil)

	require.Equal(t, []uint64{1, 2}, []uint64{graph.nodes[0].id, graph.nodes[1].id})
	require.Equal(t, []string{"main.go", "handlers.go"}, []string{graph.nodes[0].label, graph.nodes[1].label})
	require.Equal(t, []edgeModel{{from: 0, to: 1}}, graph.edges)
	require.Equal(t, 0, graph.root)
}

func TestDependencyGraphScalesConnectedFunctions(t *testing.T) {
	source := testRuntimeGraph()
	source.Static.Nodes[4] = &model.StaticNode{ID: 4, Name: "health", Syntax: model.Syntax{File: 2}}
	source.Static.Nodes[1].Out = append(source.Static.Nodes[1].Out, 4)
	source.Static.Files[2].Functions = append(source.Static.Files[2].Functions, 4)
	source.Nodes[4] = &model.Node{Static: source.Static.Nodes[4]}
	graph := New("api", DependencyGraph, source, nil)

	require.Contains(t, []uint64{graph.nodes[0].id, graph.nodes[1].id, graph.nodes[2].id, graph.nodes[3].id, graph.nodes[4].id}, uint64(99))
	require.Equal(t, "http.external", graph.nodes[4].label)
	require.Greater(t, graph.nodes[0].scale, graph.nodes[3].scale)
	for _, node := range graph.nodes {
		require.GreaterOrEqual(t, node.scale, 1.0)
		require.LessOrEqual(t, node.scale, 2.0)
	}
}

func TestGraphDisambiguatesDuplicateFunctionNames(t *testing.T) {
	source := testRuntimeGraph()
	source.Static.Nodes[2].Name = "handler"
	source.Nodes[2].Static.Name = "handler"
	graph := New("api", FunctionGraph, source, nil)

	require.Equal(t, "handler @ main.go:8", graph.nodes[1].label)
	require.Equal(t, "handler @ handlers.go:4", graph.nodes[2].label)
}

func TestGraphRendersNativeLabels(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	for _, kind := range []Kind{FunctionGraph, FileGraph, DependencyGraph} {
		graph, command := New("api", kind, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
		view := graph.View()

		require.NotNil(t, command)
		require.Equal(t, 80, lipgloss.Width(view))
		require.Equal(t, 16, lipgloss.Height(view))
		for _, node := range graph.nodes {
			require.Contains(t, view, node.label)
		}
		require.False(t, strings.ContainsRune(view, kitty.Placeholder))
		require.GreaterOrEqual(t, graph.nodes[graph.root].labelX, 0)
	}
}

func TestRasterCellSizeCapsWorkAndPreservesAspectRatio(t *testing.T) {
	width, height := rasterCellSize(10, 28)

	require.Equal(t, 7, width)
	require.Equal(t, maximumRasterHeight, height)
	require.InDelta(t, 10.0/28.0, float64(width)/float64(height), 0.01)
	width, height = rasterCellSize(15, 38)
	require.Equal(t, 8, width)
	require.Equal(t, 20, height)
}

func TestGraphDistributesNodesAcrossViewport(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))

	maximumDepth := 0
	for _, node := range graph.nodes {
		maximumDepth = max(maximumDepth, node.depth)
	}
	availableWidth := float64(graph.width - nodeWidth - int(graphInsetX))
	availableHeight := float64(graph.height - 2 - int(graphInsetY))
	for index, node := range graph.nodes {
		if index == graph.root {
			require.Equal(t, point{}, node.position)
			continue
		}
		require.InDelta(t, availableWidth*float64(node.depth)/float64(maximumDepth), node.position.x, 0.001)
		require.InDelta(t, availableHeight*float64(node.row+1)/float64(node.rowCount+1), node.position.y, 0.001)
	}
	for _, edge := range graph.edges {
		require.Equal(t, max(baseEdgeLength, distance(graph.nodes[edge.from].position, graph.nodes[edge.to].position)), edge.rest)
	}
}

func TestGraphDistributesDenseLevelWithoutStacking(t *testing.T) {
	static := &model.StaticGraph{Root: 1, Nodes: make(map[model.NodeID]*model.StaticNode)}
	static.Nodes[1] = &model.StaticNode{ID: 1, Name: "main"}
	source := &model.RuntimeGraph{
		Static: static,
		Nodes:  map[model.NodeID]*model.Node{1: {Static: static.Nodes[1]}},
	}
	for id := model.NodeID(2); id <= 36; id++ {
		static.Nodes[id] = &model.StaticNode{ID: id, Name: fmt.Sprintf("dependency-%d", id)}
		static.Nodes[1].Out = append(static.Nodes[1].Out, id)
	}
	graph, _ := New("api", DependencyGraph, source, nil).Update(visibleViewport(176, 42))

	rows := make(map[int]struct{}, len(graph.nodes)-1)
	for index, node := range graph.nodes {
		if index != graph.root {
			rows[int(math.Round(node.position.y))] = struct{}{}
		}
	}
	require.Len(t, rows, len(graph.nodes)-1)
}

func TestGraphTransmitsPNGForEveryGraphKind(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	for _, kind := range []Kind{FunctionGraph, FileGraph, DependencyGraph} {
		graph, _ := New("api", kind, testRuntimeGraph(), nil).Update(pages.ViewportMsg{X: 3, Y: 4, Width: 40, Height: 10, Visible: true})
		sequence := renderOutput(t, &graph)

		for _, option := range []string{"f=100", "q=2", "a=p", "c=40", "r=10", "z=-1", "C=1", "\x1b[5;4H"} {
			require.Contains(t, sequence, option)
		}
		require.NotContains(t, sequence, "U=1")
		require.Contains(t, sequence, "\x1b[?2026h")
		require.Contains(t, sequence, "\x1b[?2026l")
		require.Less(t, strings.Index(sequence, "f=100"), strings.Index(sequence, "a=p"))
	}
}

func TestRenderedFramePreservesLabelsForEveryGraphKind(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	for _, kind := range []Kind{FunctionGraph, FileGraph, DependencyGraph} {
		graph, _ := New("api", kind, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
		before := graph.View()
		ready := graph.upload()().(renderReadyMsg)

		graph, command := graph.Update(ready)

		require.NotNil(t, command)
		require.Equal(t, before, graph.View(), "graph kind %d changed labels after rendering", kind)
		graph, _ = graph.Update(command().(tea.RawMsg))
	}
}

func TestGraphDebugLogKeepsTerminalErrors(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	var dump bytes.Buffer
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), &dump).Update(visibleViewport(40, 10))
	output := renderOutput(t, &graph)

	require.Contains(t, output, "q=1")
	require.Contains(t, dump.String(), "nodes=3 edges=3")
	require.Contains(t, dump.String(), "frame ready")
}

func TestRenderReadyDebugSummaryOmitsImagePayload(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(40, 10))
	ready := graph.upload()().(renderReadyMsg)

	require.Contains(t, ready.DebugSummary(), "render ready sequence=1")
	require.NotContains(t, ready.DebugSummary(), "iVBOR")
}

func TestGraphAlternatesImagesAndDeletesPreviousFrame(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(40, 10))
	firstID := graph.imageIDs[0]
	first := renderOutput(t, &graph)

	require.Equal(t, firstID, graph.frontImageID)
	require.NotContains(t, first, "d=I")
	second := renderOutput(t, &graph)
	require.Equal(t, graph.imageIDs[1], graph.frontImageID)
	require.Contains(t, second, "d=I")
	require.Contains(t, second, "i="+strconv.Itoa(int(firstID)))
}

func TestHiddenGraphCancelsFramesAndDeletesImages(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(40, 10))
	_ = renderOutput(t, &graph)

	graph, command := graph.Update(pages.ViewportMsg{Width: 40, Height: 10})
	require.NotNil(t, command)
	require.False(t, graph.visible)
	require.Zero(t, graph.frontImageID)
	cleanup := command().(tea.RawMsg)
	require.Contains(t, cleanup.Msg, "d=I")
	require.Nil(t, graph.upload())
}

func TestLabelRevisionChangesOnlyAfterCrossingCell(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	revision := graph.Revision()

	graph.nodes[1].position.x += 0.01
	graph.refreshLabels(false)
	require.Equal(t, revision, graph.Revision())

	graph.nodes[1].position.x += 1
	graph.refreshLabels(false)
	require.Greater(t, graph.Revision(), revision)
}

func TestLabelCellUsesHysteresis(t *testing.T) {
	require.Equal(t, 10, stableLabelCell(10.74, 10, false))
	require.Equal(t, 11, stableLabelCell(10.76, 10, false))
	require.Equal(t, 11, stableLabelCell(10.26, 11, false))
	require.Equal(t, 10, stableLabelCell(10.24, 11, false))
}

func TestMovingLabelDoesNotRepositionOtherLabels(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	stationaryCell := graph.nodes[2].labelCellX
	graph.nodes[1].position.x = float64(graph.nodes[1].labelCellX) + 0.8 - graphInsetX
	graph.nodes[2].position.x = float64(stationaryCell) + 0.6 - graphInsetX

	graph.refreshLabels(false)

	require.Equal(t, stationaryCell, graph.nodes[2].labelCellX)
}

func TestMovingLabelCommitsWithRenderedFrame(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	_ = renderOutput(t, &graph)
	revision := graph.Revision()
	graph.dragging = 1
	graph.animating = true
	graph.nodes[1].position.x += 5

	graph, command := graph.Update(frameMsg{owner: graph.owner, generation: graph.generation})
	require.NotNil(t, command)
	require.Equal(t, revision, graph.Revision(), "label must wait for its image frame")
	ready := command().(renderReadyMsg)
	for index := range graph.nodes {
		require.Equal(t, graph.nodes[index].position, graph.nodes[index].rendered)
	}
	graph, command = graph.Update(ready)
	require.Greater(t, graph.Revision(), revision)
	committedRevision := graph.Revision()
	raw := command().(tea.RawMsg)
	graph, _ = graph.Update(raw)
	require.Equal(t, committedRevision, graph.Revision())
}

func TestGraphDrawsContinuousEdgesUnderNodes(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	canvas := graph.renderer.renderImage(renderRequest{
		nodes:      graph.nodes,
		dragging:   graph.dragging,
		width:      graph.width,
		height:     graph.height,
		cellWidth:  graph.cellWidth,
		cellHeight: graph.cellHeight,
		nodeRadius: graph.nodeRadius,
	})
	edge := graph.edges[0]
	from := pixelPoint(graph.nodes[edge.from].position, graph.cellWidth, graph.cellHeight)
	to := pixelPoint(graph.nodes[edge.to].position, graph.cellWidth, graph.cellHeight)
	middle := point{x: (from.x + to.x) / 2, y: (from.y + to.y) / 2}

	edgePixels := 0
	for y := int(middle.y) - 2; y <= int(middle.y)+2; y++ {
		for x := int(middle.x) - 2; x <= int(middle.x)+2; x++ {
			if pixelAlpha(canvas, x, y) != 0 {
				edgePixels++
			}
		}
	}
	require.Positive(t, edgePixels)
	require.NotZero(t, pixelAlpha(canvas, int(from.x), int(from.y)))
	for step := range 101 {
		ratio := float64(step) / 100
		x := int(math.Round(from.x + (to.x-from.x)*ratio))
		y := int(math.Round(from.y + (to.y-from.y)*ratio))
		connected := false
		for sampleY := y - 1; sampleY <= y+1 && !connected; sampleY++ {
			for sampleX := x - 1; sampleX <= x+1; sampleX++ {
				if pixelAlpha(canvas, sampleX, sampleY) != 0 {
					connected = true
					break
				}
			}
		}
		require.True(t, connected, "transparent gap at %.0f%%", ratio*100)
	}
}

func TestGraphUsesPalettedFrames(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	canvas := graph.renderer.renderImage(renderRequest{
		nodes:      graph.nodes,
		dragging:   graph.dragging,
		width:      graph.width,
		height:     graph.height,
		cellWidth:  graph.cellWidth,
		cellHeight: graph.cellHeight,
		nodeRadius: graph.nodeRadius,
	})

	require.Len(t, canvas.Palette, 1+3*alphaLevels)
	require.Equal(t, canvas.Bounds().Dx()*canvas.Bounds().Dy(), len(canvas.Pix))
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, canvas))
	configuration, err := png.DecodeConfig(&encoded)
	require.NoError(t, err)
	_, paletted := configuration.ColorModel.(color.Palette)
	require.True(t, paletted)
}

func TestGraphAppliesRenderBackpressure(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(200, 40))
	command := graph.upload()
	require.NotNil(t, command)
	for range 20 {
		graph.nodes[1].position.x += 0.1
		require.Nil(t, graph.upload())
	}

	require.IsType(t, renderReadyMsg{}, command())
}

func pixelAlpha(canvas image.Image, x, y int) uint32 {
	_, _, _, alpha := canvas.At(x, y).RGBA()
	return alpha
}

func TestGraphSettlesAndPinsRoot(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)

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
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)
	root := screenPoint(graph.nodes[graph.root].position)

	graph, command := graph.Update(tea.MouseClickMsg{X: int(root.x), Y: int(root.y), Button: tea.MouseLeft})

	require.Nil(t, command)
	require.Equal(t, -1, graph.dragging)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
}

func TestDraggedNodeJigglesNeighborAndReturnsToAnchor(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
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
	graph, command = graph.Update(tea.MouseMotionMsg{X: int(target.x), Y: int(target.y), Button: tea.MouseLeft})
	require.Nil(t, command, "mouse events are rendered by the frame clock")
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

func TestResizePresettlesLayout(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)
	generation := graph.generation

	graph, command := graph.Update(visibleViewport(100, 24))

	require.NotNil(t, command)
	require.Equal(t, generation, graph.generation)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
	for _, node := range graph.nodes {
		require.Equal(t, node.position, node.anchor)
	}
}

func TestGraphUsesUniqueImageIDs(t *testing.T) {
	first := New("api", FunctionGraph, testRuntimeGraph(), nil)
	second := New("worker", FunctionGraph, testRuntimeGraph(), nil)
	require.NotZero(t, first.imageIDs[0])
	require.NotEqual(t, first.imageIDs, second.imageIDs)
}

func TestUnsupportedTerminalHasTextFallback(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "Alacritty")
	t.Setenv("KITTY_WINDOW_ID", "")
	graph, command := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(60, 12))
	view := graph.View()

	require.Nil(t, command)
	require.Contains(t, view, "requires Ghostty or Kitty")
	require.False(t, strings.ContainsRune(view, kitty.Placeholder))
}

func TestEmptyGraphHasTextFallback(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, command := New("api", FunctionGraph, nil, nil).Update(visibleViewport(60, 12))

	require.Nil(t, command)
	require.Contains(t, graph.View(), "No project functions found")
}

func TestGraphIgnoresFramesForAnotherServer(t *testing.T) {
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	before := graph.nodes[1].position

	graph, command := graph.Update(frameMsg{owner: "worker", generation: graph.generation})

	require.Nil(t, command)
	require.Equal(t, before, graph.nodes[1].position)
}

func BenchmarkGraphUpload(b *testing.B) {
	b.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	wireBytes := len(renderOutput(b, &graph))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = renderOutput(b, &graph)
	}
	b.ReportMetric(float64(wireBytes), "wire-B/frame")
}

func BenchmarkGraphLargeUpload(b *testing.B) {
	b.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(176, 42))
	graph.cellWidth, graph.cellHeight = rasterCellSize(15, 38)
	graph.nodeRadius = max(3, float64(graph.cellWidth)*0.65)
	wireBytes := len(renderOutput(b, &graph))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = renderOutput(b, &graph)
	}
	b.ReportMetric(float64(wireBytes), "wire-B/frame")
}

func BenchmarkGraphBuild(b *testing.B) {
	source := testRuntimeGraph()
	b.ReportAllocs()
	for b.Loop() {
		_ = New("api", FunctionGraph, source, nil)
	}
}

func BenchmarkGraphView(b *testing.B) {
	b.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	b.ReportAllocs()
	for b.Loop() {
		if graph.View() == "" {
			b.Fatal("empty graph view")
		}
	}
}

func visibleViewport(width, height int) pages.ViewportMsg {
	return pages.ViewportMsg{Width: width, Height: height, Visible: true}
}

func renderOutput(tb testing.TB, graph *Model) string {
	tb.Helper()
	ready, ok := graph.upload()().(renderReadyMsg)
	require.True(tb, ok)
	next, command := graph.Update(ready)
	*graph = next
	require.NotNil(tb, command)
	raw, ok := command().(tea.RawMsg)
	require.True(tb, ok)
	next, _ = graph.Update(raw)
	*graph = next
	return raw.Msg.(frameOutput).String()
}

func settleLayout(graph Model) Model {
	for frame := 0; frame < maximumReturnFrames+1 && graph.animating; frame++ {
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
			99:  {ID: 99, Name: "external", Pkg: 9},
			100: {ID: 100, Name: "root", Out: []model.NodeID{1}},
		},
		Files: map[model.FileID]*model.StaticFile{
			1: {ID: 1, Path: "/project/main.go", Functions: []model.NodeID{1, 2}, Calls: []model.FileID{2}},
			2: {ID: 2, Path: "/project/handlers.go", Functions: []model.NodeID{3}, CalledBy: []model.FileID{1}},
		},
		Packages: map[model.PackageID]*model.Package{
			9: {Name: "http", Path: "net/http"},
		},
	}
	return &model.RuntimeGraph{
		Static: static,
		Nodes: map[model.NodeID]*model.Node{
			1: {Static: static.Nodes[1]},
			2: {Static: static.Nodes[2]},
			3: {Static: static.Nodes[3]},
		},
		Files: map[model.FileID]*model.File{
			1: {Static: static.Files[1]},
			2: {Static: static.Files[2]},
		},
	}
}
