package servers

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/x/ansi/kitty"
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
	page, command := server.Update(tea.WindowSizeMsg{Width: 60, Height: 12})
	require.NotNil(t, command)
	page, command = page.Update(tea.RawMsg{})
	require.Nil(t, command)

	view := page.(Model).View()
	require.Equal(t, 60, lipgloss.Width(view.Content))
	require.Equal(t, 12, lipgloss.Height(view.Content))
	require.Equal(t, tea.MouseModeCellMotion, view.MouseMode)
	require.Contains(t, view.Content, string(kitty.Placeholder))
	require.Contains(t, view.Content, "●")
	require.NotContains(t, view.Content, "drag nodes")
}

func TestServerSwitchesViewsFromPager(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	static := &model.StaticGraph{Root: 1, Nodes: map[model.NodeID]*model.StaticNode{1: {ID: 1, Name: "main"}}}
	server := New(&engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph:   &model.RuntimeGraph{Static: static, Nodes: map[model.NodeID]*model.Node{1: {Static: static.Nodes[1]}}},
	}, nil)
	page, _ := server.Update(tea.WindowSizeMsg{Width: 21, Height: 8})
	page, _ = page.Update(tea.MouseClickMsg{X: 10, Y: 7, Button: tea.MouseLeft})

	require.Contains(t, page.(Model).View().Content, "Runtime")
}
