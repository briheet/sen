package servers

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/model"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/stretchr/testify/require"
)

func TestServerRendersGraphViewport(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	static := &model.StaticGraph{
		Root:  1,
		Nodes: map[model.NodeID]*model.StaticNode{1: {ID: 1, Name: "main"}},
	}
	server := New(&engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph: &model.RuntimeGraph{
			Static: static,
			Nodes:  map[model.NodeID]*model.Node{1: {Static: static.Nodes[1]}},
		},
	}, nil)
	page, command := server.Update(pages.ViewportMsg{X: 1, Y: 2, Width: 60, Height: 12, Visible: true})
	require.NotNil(t, command)

	view := page.(Model).View()
	require.Equal(t, 60, lipgloss.Width(view.Content))
	require.Equal(t, 12, lipgloss.Height(view.Content))
	require.Equal(t, tea.MouseModeCellMotion, view.MouseMode)
	require.Contains(t, view.Content, "main")
	require.Contains(t, view.Content, "●")
	require.NotContains(t, view.Content, "drag nodes")
}

func TestServerPreservesFunctionLabels(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	static := &model.StaticGraph{
		Root: 1,
		Nodes: map[model.NodeID]*model.StaticNode{
			1: {ID: 1, Name: "main", Out: []model.NodeID{2}},
			2: {ID: 2, Name: "routes", Out: []model.NodeID{3}},
			3: {ID: 3, Name: "handler"},
		},
	}
	server := New(&engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph: &model.RuntimeGraph{
			Static: static,
			Nodes: map[model.NodeID]*model.Node{
				1: {Static: static.Nodes[1]},
				2: {Static: static.Nodes[2]},
				3: {Static: static.Nodes[3]},
			},
		},
	}, nil)
	page, _ := server.Update(pages.ViewportMsg{Width: 80, Height: 16, Visible: true})
	view := page.View().Content

	require.Contains(t, view, "main")
	require.Contains(t, view, "routes")
	require.Contains(t, view, "handler")
}

func TestServerSwitchesViewsFromPager(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	static := &model.StaticGraph{
		Root: 1,
		Nodes: map[model.NodeID]*model.StaticNode{
			1: {ID: 1, Name: "main", Syntax: model.Syntax{File: 1}},
		},
		Files: map[model.FileID]*model.StaticFile{
			1: {ID: 1, Path: "/project/main.go", Functions: []model.NodeID{1}},
		},
	}
	server := New(&engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph: &model.RuntimeGraph{
			Static: static,
			Nodes:  map[model.NodeID]*model.Node{1: {Static: static.Nodes[1]}},
			Files:  map[model.FileID]*model.File{1: {Static: static.Files[1]}},
		},
	}, nil)
	page, _ := server.Update(pages.ViewportMsg{Width: 21, Height: 8, Visible: true})
	require.Equal(t, 2, page.(Model).pager.TotalPages)
	page, command := page.Update(tea.MouseClickMsg{X: 10, Y: 7, Button: tea.MouseLeft})

	require.Equal(t, 0, page.(Model).pager.Page)
	require.Equal(t, 1, page.(Model).pending)
	require.NotNil(t, command)
	page, command = page.Update(switchGraphMsg{service: "api", page: 1})
	require.Equal(t, 1, page.(Model).pager.Page)
	require.Contains(t, page.(Model).View().Content, "main.go")
	require.NotNil(t, command)
}

func TestServerSwitchesViewsFromShiftKeybindings(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	static := &model.StaticGraph{
		Root: 1,
		Nodes: map[model.NodeID]*model.StaticNode{
			1: {ID: 1, Name: "main", Syntax: model.Syntax{File: 1}},
		},
		Files: map[model.FileID]*model.StaticFile{
			1: {ID: 1, Path: "/project/main.go", Functions: []model.NodeID{1}},
		},
	}
	server := New(&engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph: &model.RuntimeGraph{
			Static: static,
			Nodes:  map[model.NodeID]*model.Node{1: {Static: static.Nodes[1]}},
			Files:  map[model.FileID]*model.File{1: {Static: static.Files[1]}},
		},
	}, nil)
	page, _ := server.Update(pages.ViewportMsg{Width: 21, Height: 8, Visible: true})

	page, command := page.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModShift, Text: "L"})
	require.Equal(t, 0, page.(Model).pager.Page)
	require.Equal(t, 1, page.(Model).pending)
	require.NotNil(t, command)

	page, command = page.Update(switchGraphMsg{service: "api", page: 1})
	require.Equal(t, 1, page.(Model).pager.Page)
	require.Contains(t, page.(Model).View().Content, "main.go")
	require.NotNil(t, command)

	page, command = page.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModShift, Text: "H"})
	require.Equal(t, 1, page.(Model).pager.Page)
	require.Equal(t, 0, page.(Model).pending)
	require.NotNil(t, command)

	page, _ = page.Update(switchGraphMsg{service: "api", page: 0})
	require.Equal(t, 0, page.(Model).pager.Page)
}

func TestServerTogglesMetricsOverlay(t *testing.T) {
	t.Setenv("TERM", "dumb")
	static := &model.StaticGraph{
		Root:  1,
		Nodes: map[model.NodeID]*model.StaticNode{1: {ID: 1, Name: "main"}},
	}
	server := New(&engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph: &model.RuntimeGraph{
			Static: static,
			Nodes:  map[model.NodeID]*model.Node{1: {Static: static.Nodes[1]}},
		},
	}, nil)
	page, _ := server.Update(pages.ViewportMsg{Width: 80, Height: 18, Visible: true})
	revision := page.Revision()

	page, _ = page.Update(tea.KeyPressMsg{Code: 'm', Mod: tea.ModShift, Text: "M"})
	view := page.View().Content
	require.Contains(t, view, "live heap")
	require.Greater(t, page.Revision(), revision)

	page, _ = page.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotContains(t, page.View().Content, "live heap")

	page, _ = page.Update(tea.KeyPressMsg{Code: 'm', Mod: tea.ModShift, Text: "M"})
	require.Contains(t, page.View().Content, "live heap")

	page, _ = page.Update(tea.KeyPressMsg{Code: 'm', Mod: tea.ModShift, Text: "M"})
	require.NotContains(t, page.View().Content, "live heap")
}
