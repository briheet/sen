package tui

import (
	"bytes"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	domain "github.com/briheet/sen/internal/model"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/stretchr/testify/require"
)

func TestProgramNavigatesPages(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	engines := []*engine.Engine{
		{Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo}},
		{Service: config.Service{Name: "worker", Type: config.ServiceTypeServer, Lang: config.ServiceLangNode}},
	}
	services := []config.Service{
		engines[0].Service,
		{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis, Address: "localhost:6379"},
		engines[1].Service,
	}
	tm := teatest.NewTestModel(t, initialModel(engines, services, "/tmp/engine.log", nil),
		teatest.WithInitialTermSize(80, 24),
	)
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("api")) &&
			bytes.Contains(output, []byte("more"))
	})
	tm.Send(tea.MouseClickMsg{X: 5, Y: 12, Button: tea.MouseLeft})
	tm.Send(tea.MouseMotionMsg{X: 15, Y: 14, Button: tea.MouseLeft})
	tm.Send(tea.MouseReleaseMsg{X: 15, Y: 14, Button: tea.MouseLeft})
	tm.Type("?")
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("less"))
	})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyRight})
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("cache (redis)"))
	})
	tm.Type("q")

	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second)).(model)
	require.Equal(t, "cache", final.ctx.ActivePage())
	require.Equal(t, 80, final.width)
	require.Equal(t, 24, final.height)
}

func TestProgramQuitsWhenEnginesFinish(t *testing.T) {
	tm := teatest.NewTestModel(t, initialModel(nil, nil, "", nil))
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	tm.Send(enginesDoneMsg{})
	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))

	require.IsType(t, model{}, final)
}

func TestProgramEmitsKittyGraph(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("KITTY_WINDOW_ID", "")
	engines := []*engine.Engine{{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph:   programRuntimeGraph(),
	}}
	tm := teatest.NewTestModel(t, initialModel(engines, []config.Service{engines[0].Service}, "", nil),
		teatest.WithInitialTermSize(60, 20),
	)
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	var terminalOutput []byte
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		terminalOutput = append(terminalOutput[:0], output...)
		altScreen := bytes.Index(output, []byte("\x1b[?1049h"))
		return altScreen >= 0 &&
			bytes.LastIndex(output, []byte("\x1b_Gf=100")) > altScreen &&
			bytes.LastIndex(output, []byte("a=p,z=-1")) > altScreen &&
			bytes.Index(output, []byte("main")) > altScreen
	})
	require.Less(t,
		bytes.Index(terminalOutput, []byte("\x1b[?1049h")),
		bytes.LastIndex(terminalOutput, []byte("\x1b_Gf=100")),
	)
}

func TestProgramDeletesGraphWhenSwitchingServices(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "ghostty")
	server := &engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph:   programRuntimeGraph(),
	}
	services := []config.Service{
		server.Service,
		{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis},
	}
	tm := teatest.NewTestModel(t, initialModel([]*engine.Engine{server}, services, "", nil),
		teatest.WithInitialTermSize(60, 20),
	)
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("a=p,z=-1"))
	})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyRight})
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("cache (redis)")) &&
			bytes.Contains(output, []byte("a=d")) &&
			bytes.Contains(output, []byte("d=I"))
	})
}

func programRuntimeGraph() *domain.RuntimeGraph {
	static := &domain.StaticGraph{
		Root: 1,
		Nodes: map[domain.NodeID]*domain.StaticNode{
			1: {ID: 1, Name: "main", Out: []domain.NodeID{2}},
			2: {ID: 2, Name: "handler", In: []domain.NodeID{1}},
		},
	}
	return &domain.RuntimeGraph{
		Static: static,
		Nodes: map[domain.NodeID]*domain.Node{
			1: {Static: static.Nodes[1]},
			2: {Static: static.Nodes[2]},
		},
	}
}
