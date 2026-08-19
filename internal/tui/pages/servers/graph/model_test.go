package graph

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/stretchr/testify/require"
)

func TestGraphBuildsProjectFunctionsAndCalls(t *testing.T) {
	graph := New("api", FunctionGraph, testRuntimeGraph(), nil)

	require.Equal(t, []uint64{1, 2, 3, 99}, []uint64{
		graph.nodes[0].id,
		graph.nodes[1].id,
		graph.nodes[2].id,
		graph.nodes[3].id,
	})
	require.Equal(t, []string{"main", "routes", "handler", "http.external"}, []string{
		graph.nodes[0].label,
		graph.nodes[1].label,
		graph.nodes[2].label,
		graph.nodes[3].label,
	})
	require.Equal(t, 0, graph.root)
	distance := layoutFor(FunctionGraph).linkDistance
	require.Equal(t, []edgeModel{
		{from: 0, to: 1, distance: distance},
		{from: 0, to: 2, distance: distance},
		{from: 0, to: 3, distance: distance},
		{from: 1, to: 2, distance: distance},
	}, graph.edges)
}

func TestGraphBuildsProjectFilesAndCalls(t *testing.T) {
	graph := New("api", FileGraph, testRuntimeGraph(), nil)

	require.Equal(t, []uint64{1, 2}, []uint64{graph.nodes[0].id, graph.nodes[1].id})
	require.Equal(t, []string{"main.go", "handlers.go"}, []string{graph.nodes[0].label, graph.nodes[1].label})
	require.Equal(t, []edgeModel{{from: 0, to: 1, distance: 96}}, graph.edges)
	require.Equal(t, 0, graph.root)
}

func TestFunctionGraphScalesConnectedFunctions(t *testing.T) {
	source := testRuntimeGraph()
	source.Static.Nodes[4] = &model.StaticNode{ID: 4, Name: "health", Syntax: model.Syntax{File: 2}}
	source.Static.Nodes[1].Out = append(source.Static.Nodes[1].Out, 4)
	source.Static.Files[2].Functions = append(source.Static.Files[2].Functions, 4)
	source.Nodes[4] = &model.Node{Static: source.Static.Nodes[4]}
	graph := New("api", FunctionGraph, source, nil)

	require.Contains(t, []uint64{graph.nodes[0].id, graph.nodes[1].id, graph.nodes[2].id, graph.nodes[3].id, graph.nodes[4].id}, uint64(99))
	require.Equal(t, "http.external", graph.nodes[4].label)
	require.Greater(t, graph.nodes[0].scale, graph.nodes[3].scale)
	for _, node := range graph.nodes {
		require.GreaterOrEqual(t, node.scale, 1.0)
		require.LessOrEqual(t, node.scale, 2.0)
	}
}

func TestTelemetryHighlightsObservedTracePath(t *testing.T) {
	graph := New("api", FunctionGraph, testRuntimeGraph(), nil)
	graph.graphics = false
	graph, command := graph.Update(TelemetryMsg{
		Nodes:     map[model.NodeID]int64{1: 4, 2: 2},
		NodeEdges: map[model.NodeEdge]int64{{From: 1, To: 2}: 2},
	})

	require.Nil(t, command)
	require.True(t, graph.nodes[0].active)
	require.True(t, graph.nodes[1].active)
	require.False(t, graph.nodes[2].active)
	request := graphRenderRequest(graph)
	require.Equal(t, 2, edgeClass(&request, graph.edges[0]))
	require.Equal(t, 1, edgeClass(&request, graph.edges[1]))
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
	for _, kind := range []Kind{FunctionGraph, FileGraph} {
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

func TestGraphLayoutIsDeterministicAndUsesUniformLinks(t *testing.T) {
	first := New("api", FunctionGraph, testRuntimeGraph(), nil)
	second := New("api", FunctionGraph, testRuntimeGraph(), nil)

	require.Equal(t, point{}, first.nodes[first.root].position)
	require.False(t, first.nodes[first.root].fixed)
	for index, node := range first.nodes {
		require.Equal(t, second.nodes[index].position, node.position)
		require.False(t, math.IsNaN(node.position.x))
		require.False(t, math.IsNaN(node.position.y))
	}
	for _, edge := range first.edges {
		require.Equal(t, layoutFor(FunctionGraph).linkDistance, edge.distance)
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
	graph, _ := New("api", FunctionGraph, source, nil).Update(visibleViewport(176, 42))

	positions := make(map[point]struct{}, len(graph.nodes))
	for _, node := range graph.nodes {
		require.False(t, math.IsNaN(node.position.x))
		require.False(t, math.IsNaN(node.position.y))
		positions[node.position] = struct{}{}
	}
	require.Len(t, positions, len(graph.nodes))
}

func TestGraphTransmitsPNGForEveryGraphKind(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	for _, kind := range []Kind{FunctionGraph, FileGraph} {
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
	for _, kind := range []Kind{FunctionGraph, FileGraph} {
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
	require.Contains(t, dump.String(), "nodes=4 edges=4")
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

	screen := graph.camera.worldToScreen(graph.nodes[1].position)
	screen.x += float64(graph.cellWidth) * 2
	graph.nodes[1].position = graph.camera.screenToWorld(screen)
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
	moving := graph.camera.worldToScreen(graph.nodes[1].position)
	stationary := graph.camera.worldToScreen(graph.nodes[2].position)
	moving.x += float64(graph.cellWidth)
	stationary.x += float64(graph.cellWidth) * 0.6
	graph.nodes[1].position = graph.camera.screenToWorld(moving)
	graph.nodes[2].position = graph.camera.screenToWorld(stationary)

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
	request := graphRenderRequest(graph)
	canvas := renderCanvas(graph.renderer, request)
	edge := graph.edges[0]
	from := graph.camera.worldToScreen(graph.nodes[edge.from].position)
	to := graph.camera.worldToScreen(graph.nodes[edge.to].position)
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
	canvas := renderCanvas(graph.renderer, graphRenderRequest(graph))

	require.Len(t, canvas.Palette, 1+3*alphaLevels)
	require.Equal(t, canvas.Bounds().Dx()*canvas.Bounds().Dy(), len(canvas.Pix))
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, canvas))
	configuration, err := png.DecodeConfig(&encoded)
	require.NoError(t, err)
	_, paletted := configuration.ColorModel.(color.Palette)
	require.True(t, paletted)
}

func TestDirectCapsulesStayContinuousAtEverySlope(t *testing.T) {
	palette := newGraphPalette(styles.Zakura)
	for _, segment := range [][2]point{
		{{x: 5, y: 5}, {x: 70, y: 5}},
		{{x: 5, y: 5}, {x: 5, y: 70}},
		{{x: 5, y: 8}, {x: 70, y: 63}},
		{{x: 8, y: 70}, {x: 63, y: 5}},
	} {
		canvas := image.NewPaletted(image.Rect(0, 0, 80, 80), palette.colors)
		drawCapsule(canvas, &palette.active, segment[0], segment[1], edgeWidth, 1)
		for step := range 101 {
			ratio := float64(step) / 100
			x := int(math.Round(segment[0].x + (segment[1].x-segment[0].x)*ratio))
			y := int(math.Round(segment[0].y + (segment[1].y-segment[0].y)*ratio))
			connected := false
			for sampleY := y - 1; sampleY <= y+1 && !connected; sampleY++ {
				for sampleX := x - 1; sampleX <= x+1; sampleX++ {
					if pixelAlpha(canvas, sampleX, sampleY) != 0 {
						connected = true
						break
					}
				}
			}
			require.True(t, connected, "transparent gap at %.0f%% for %v", ratio*100, segment)
		}
	}
}

func TestKittyChunkWriterStreamsPayload(t *testing.T) {
	payload := bytes.Repeat([]byte("a"), kitty.MaxChunkSize*2+17)
	options := &kitty.Options{Action: kitty.Transmit, Format: kitty.PNG, Quiet: 2}
	var output bytes.Buffer
	var writer kittyChunkWriter
	writer.reset(&output, options)

	written, err := writer.Write(payload[:3000])
	require.NoError(t, err)
	require.Equal(t, 3000, written)
	written, err = writer.Write(payload[3000:])
	require.NoError(t, err)
	require.Equal(t, len(payload)-3000, written)
	writer.close()

	require.Contains(t, output.String(), "m=1")
	require.Contains(t, output.String(), "m=0")
	require.Equal(t, payload, kittyPayload(output.String()))
}

func TestVisualHashIgnoresSubpixelMotion(t *testing.T) {
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	screen := point{x: 100, y: 100}
	graph.nodes[1].position = graph.camera.screenToWorld(screen)
	baseline := graph.visualHash()

	screen.x += 0.1
	graph.nodes[1].position = graph.camera.screenToWorld(screen)
	require.Equal(t, baseline, graph.visualHash())

	screen.x += 0.2
	graph.nodes[1].position = graph.camera.screenToWorld(screen)
	require.NotEqual(t, baseline, graph.visualHash())
}

func TestLabelScratchIsReused(t *testing.T) {
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph.refreshLabels(true)
	cells := &graph.labelScratch.cells[0]
	occupied := &graph.labelScratch.occupied[0]

	graph.refreshLabels(true)

	require.Equal(t, cells, &graph.labelScratch.cells[0])
	require.Equal(t, occupied, &graph.labelScratch.occupied[0])
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
	if !image.Pt(x, y).In(canvas.Bounds()) {
		return 0
	}
	_, _, _, alpha := canvas.At(x, y).RGBA()
	return alpha
}

func TestGraphStartsWithRootAtOrigin(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)

	require.False(t, graph.animating)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
	require.False(t, graph.nodes[graph.root].fixed)
	for _, node := range graph.nodes {
		require.False(t, math.IsNaN(node.position.x))
		require.False(t, math.IsNaN(node.position.y))
	}
}

func TestRootCanBeDragged(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)
	root := mouseCell(graph, graph.nodes[graph.root].position)
	graph, command := graph.Update(tea.MouseClickMsg{X: root[0], Y: root[1], Button: tea.MouseLeft})
	require.Nil(t, command)
	graph, command = graph.Update(tea.MouseMotionMsg{X: root[0] + 5, Y: root[1] + 2, Button: tea.MouseLeft})

	require.NotNil(t, command)
	require.Equal(t, graph.root, graph.dragging)
	require.NotEqual(t, point{}, graph.nodes[graph.root].position)

	graph, _ = graph.Update(tea.MouseReleaseMsg{X: root[0] + 5, Y: root[1] + 2, Button: tea.MouseLeft})
	dropped := graph.nodes[graph.root].position
	for range 5 {
		graph, _ = graph.Update(frameMsg{owner: "api", generation: graph.generation})
	}
	require.False(t, graph.nodes[graph.root].fixed)
	require.NotEqual(t, dropped, graph.nodes[graph.root].position)
}

func TestDraggedNodeSettlesFromDroppedPosition(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)
	dragged := 1
	neighbor := 2
	original := graph.nodes[dragged].position
	neighborBefore := graph.nodes[neighbor].position
	draggedPosition := mouseCell(graph, graph.nodes[dragged].position)
	target := [2]int{draggedPosition[0] + 10, draggedPosition[1] + 3}
	click := tea.MouseClickMsg{
		X:      draggedPosition[0],
		Y:      draggedPosition[1],
		Button: tea.MouseLeft,
	}

	graph, command := graph.Update(click)
	require.Nil(t, command)
	graph, command = graph.Update(tea.MouseMotionMsg{X: target[0], Y: target[1], Button: tea.MouseLeft})
	require.Equal(t, dragged, graph.dragging)
	require.NotNil(t, command)
	for range 5 {
		graph, _ = graph.Update(frameMsg{owner: "api", generation: graph.generation})
	}
	require.NotEqual(t, neighborBefore, graph.nodes[neighbor].position)

	graph, _ = graph.Update(tea.MouseReleaseMsg{X: target[0], Y: target[1], Button: tea.MouseLeft})
	dropped := graph.nodes[dragged].position
	for frame := 0; frame < maximumLayoutTicks+1 && graph.animating; frame++ {
		graph, _ = graph.Update(frameMsg{owner: "api", generation: graph.generation})
	}
	require.False(t, graph.animating)
	require.NotEqual(t, original, graph.nodes[dragged].position)
	require.Less(t, distance(dropped, graph.nodes[dragged].position), distance(original, dropped))
	require.False(t, graph.nodes[graph.root].fixed)
}

func TestResizePresettlesLayout(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	graph = settleLayout(graph)
	generation := graph.generation
	positions := make([]point, len(graph.nodes))
	for index := range graph.nodes {
		positions[index] = graph.nodes[index].position
	}

	graph, command := graph.Update(visibleViewport(100, 24))

	require.NotNil(t, command)
	require.Equal(t, generation, graph.generation)
	require.Equal(t, point{x: 0, y: 0}, graph.nodes[graph.root].position)
	for index, node := range graph.nodes {
		require.Equal(t, positions[index], node.position)
	}
}

func TestCameraZoomKeepsWorldPointUnderCursor(t *testing.T) {
	graph, _ := New("api", FileGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	cursor := point{x: 120, y: 80}
	before := graph.camera.screenToWorld(cursor)

	graph.camera.zoomAt(cursor, zoomStep)

	require.InDelta(t, before.x, graph.camera.screenToWorld(cursor).x, 0.001)
	require.InDelta(t, before.y, graph.camera.screenToWorld(cursor).y, 0.001)
	require.True(t, graph.camera.manual)
}

func TestGraphWheelZoomsAndResetFits(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FileGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	initial := graph.camera.zoom

	graph, command := graph.Update(tea.MouseWheelMsg{X: 20, Y: 8, Button: tea.MouseWheelUp})
	require.NotNil(t, command)
	require.Greater(t, graph.camera.zoom, initial)
	require.True(t, graph.camera.manual)

	graph.renderPending = false
	graph, command = graph.Update(tea.KeyPressMsg{Code: '0'})
	require.NotNil(t, command)
	require.False(t, graph.camera.manual)
}

func TestClickSelectsNodeAndFocusesNeighbors(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	source := testRuntimeGraph()
	source.Static.Nodes[4] = &model.StaticNode{ID: 4, Name: "unused", Syntax: model.Syntax{File: 2}}
	source.Nodes[4] = &model.Node{Static: source.Static.Nodes[4]}
	graph, _ := New("api", FunctionGraph, source, nil).Update(visibleViewport(80, 20))
	cell := mouseCell(graph, graph.nodes[graph.root].position)

	graph, _ = graph.Update(tea.MouseClickMsg{X: cell[0], Y: cell[1], Button: tea.MouseLeft})
	graph, command := graph.Update(tea.MouseReleaseMsg{X: cell[0], Y: cell[1], Button: tea.MouseLeft})

	require.NotNil(t, command)
	require.Equal(t, graph.root, graph.selected)
	unused := slices.IndexFunc(graph.nodes, func(node node) bool { return node.id == 4 })
	require.GreaterOrEqual(t, unused, 0)
	request := graphRenderRequest(graph)
	renderer := newRenderer("test", nil)
	renderer.classifyNodes(&request)
	require.Equal(t, uint8(2), renderer.nodeClass[graph.root])
	require.Zero(t, renderer.nodeClass[unused])

	graph.renderPending = false
	graph, command = graph.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, command)
	require.Equal(t, -1, graph.selected)
}

func TestEmptyCanvasDragPans(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FileGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 20))
	empty := [2]int{-1, -1}
	for y := 0; y < graph.height && empty[0] < 0; y++ {
		for x := 0; x < graph.width; x++ {
			if graph.hitNode(graph.mousePixel(x, y)) < 0 {
				empty = [2]int{x, y}
				break
			}
		}
	}
	require.GreaterOrEqual(t, empty[0], 0)
	before := graph.camera.center

	graph, _ = graph.Update(tea.MouseClickMsg{X: empty[0], Y: empty[1], Button: tea.MouseLeft})
	graph, command := graph.Update(tea.MouseMotionMsg{X: empty[0] + 2, Y: empty[1] + 1, Button: tea.MouseLeft})

	require.NotNil(t, command)
	require.NotEqual(t, before, graph.camera.center)
	require.True(t, graph.camera.manual)
}

func TestVisibleLabelsDoNotOverlap(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(30, 8))
	occupied := make(map[int]struct{})

	for _, node := range graph.nodes {
		if !node.labelVisible {
			continue
		}
		for offset := range len([]rune(node.label)) {
			cell := node.labelY*graph.width + node.labelX + offset
			_, exists := occupied[cell]
			require.False(t, exists)
			occupied[cell] = struct{}{}
		}
	}
}

func TestBarnesHutLayoutProducesFinitePositions(t *testing.T) {
	graph := New("api", FunctionGraph, linearRuntimeGraph(barnesHutThreshold+1), nil)

	require.Len(t, graph.nodes, barnesHutThreshold+1)
	for _, node := range graph.nodes {
		require.False(t, math.IsNaN(node.position.x))
		require.False(t, math.IsInf(node.position.x, 0))
		require.False(t, math.IsNaN(node.position.y))
		require.False(t, math.IsInf(node.position.y, 0))
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

func TestGraphObscuredStateMutesLabels(t *testing.T) {
	graph, _ := New("api", FunctionGraph, testRuntimeGraph(), nil).Update(visibleViewport(80, 16))
	revision := graph.Revision()

	graph, _ = graph.Update(pages.ObscuredMsg{Obscured: true})

	require.True(t, graph.obscured)
	require.Greater(t, graph.Revision(), revision)
}

func TestSoftenRenderSurfaceSpreadsAndDimsPixels(t *testing.T) {
	palette := newGraphPalette(styles.Zakura)
	canvas := image.NewPaletted(image.Rect(0, 0, 25, 25), palette.colors)
	for y := 10; y < 15; y++ {
		for x := 10; x < 15; x++ {
			canvas.SetColorIndex(x, y, palette.hot[alphaLevels])
		}
	}
	surface := &renderSurface{canvas: canvas}

	softenRenderSurface(surface, &palette)

	nonTransparent := 0
	var peakAlpha uint32
	for _, index := range canvas.Pix {
		if index == 0 {
			continue
		}
		nonTransparent++
		_, _, _, alpha := canvas.Palette[index].RGBA()
		peakAlpha = max(peakAlpha, alpha)
		require.LessOrEqual(t, index, palette.idle[alphaLevels])
	}
	require.Greater(t, nonTransparent, 25)
	require.Less(t, peakAlpha, uint32(0xffff))
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

func BenchmarkForceLayout(b *testing.B) {
	for _, size := range []int{36, 256, 1000} {
		source := linearRuntimeGraph(size)
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = New("api", FunctionGraph, source, nil)
			}
		})
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
	for frame := 0; frame < maximumLayoutTicks+1 && graph.animating; frame++ {
		graph, _ = graph.Update(frameMsg{owner: graph.owner, generation: graph.generation})
	}
	return graph
}

func graphRenderRequest(graph Model) renderRequest {
	return renderRequest{
		nodes:      graph.nodes,
		edges:      graph.edges,
		camera:     graph.camera,
		dragging:   graph.dragging,
		selected:   graph.selected,
		hovered:    graph.hovered,
		width:      graph.width,
		height:     graph.height,
		cellWidth:  graph.cellWidth,
		cellHeight: graph.cellHeight,
		nodeRadius: graph.nodeRadius,
	}
}

func renderCanvas(renderer *renderer, request renderRequest) *image.Paletted {
	bounds := image.Rect(0, 0, request.width*request.cellWidth, request.height*request.cellHeight)
	canvas := image.NewPaletted(bounds, renderer.palette.colors)
	renderer.renderImage(&request, canvas)
	return canvas
}

func kittyPayload(sequence string) []byte {
	var payload bytes.Buffer
	for _, command := range strings.Split(sequence, "\x1b_G")[1:] {
		end := strings.Index(command, "\x1b\\")
		if end < 0 {
			continue
		}
		if separator := strings.IndexByte(command[:end], ';'); separator >= 0 {
			payload.WriteString(command[separator+1 : end])
		}
	}
	return payload.Bytes()
}

func mouseCell(graph Model, world point) [2]int {
	screen := graph.camera.worldToScreen(world)
	return [2]int{
		int(math.Floor(screen.x / float64(graph.cellWidth))),
		int(math.Floor(screen.y / float64(graph.cellHeight))),
	}
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

func linearRuntimeGraph(size int) *model.RuntimeGraph {
	static := &model.StaticGraph{Root: 1, Nodes: make(map[model.NodeID]*model.StaticNode, size)}
	source := &model.RuntimeGraph{Static: static, Nodes: make(map[model.NodeID]*model.Node, size)}
	for id := model.NodeID(1); id <= model.NodeID(size); id++ {
		node := &model.StaticNode{ID: id, Name: fmt.Sprintf("function-%d", id)}
		if id < model.NodeID(size) {
			node.Out = []model.NodeID{id + 1}
		}
		static.Nodes[id] = node
		source.Nodes[id] = &model.Node{Static: node}
	}
	return source
}
