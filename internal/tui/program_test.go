package tui

import (
	"bytes"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	postgresanalysis "github.com/briheet/sen/internal/adapters/postgres/analysis"
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
			bytes.Contains(output, []byte("? help"))
	})
	tm.Send(tea.MouseClickMsg{X: 5, Y: 12, Button: tea.MouseLeft})
	tm.Send(tea.MouseMotionMsg{X: 15, Y: 14, Button: tea.MouseLeft})
	tm.Send(tea.MouseReleaseMsg{X: 15, Y: 14, Button: tea.MouseLeft})
	tm.Type("?")
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("HELP")) &&
			bytes.Contains(output, []byte("move graph"))
	})
	tm.Type("?")
	tm.Type("l")
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("cache (redis)"))
	})
	tm.Type("h")
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("api"))
	})
	tm.Type("q")

	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second)).(*model)
	require.Equal(t, "api", final.ctx.ActivePage())
	require.Equal(t, 80, final.width)
	require.Equal(t, 24, final.height)
}

func TestProgramQuitsWhenEnginesFinish(t *testing.T) {
	tm := teatest.NewTestModel(t, initialModel(nil, nil, "", nil))
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	tm.Send(enginesDoneMsg{})
	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))

	require.IsType(t, &model{}, final)
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

func TestProgramPaintsLabelsWhenSwitchingGraphKinds(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "ghostty")
	target := &engine.Engine{
		Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo},
		Graph:   programRuntimeGraph(),
	}
	tm := teatest.NewTestModel(t, initialModel([]*engine.Engine{target}, []config.Service{target.Service}, "", nil),
		teatest.WithInitialTermSize(60, 20),
	)
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		placement := bytes.LastIndex(output, []byte("a=p,z=-1"))
		main := bytes.Index(output, []byte("main"))
		dependency := bytes.Index(output, []byte("http.Serve"))
		return placement >= 0 && main >= 0 && dependency >= 0 && main < placement && dependency < placement
	})

	// The pager occupies the final row of the server viewport.
	tm.Send(tea.MouseClickMsg{X: 30, Y: 18, Button: tea.MouseLeft})
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		placement := bytes.LastIndex(output, []byte("a=p,z=-1"))
		label := bytes.Index(output, []byte("main.go"))
		return label >= 0 && placement > label
	})
}

func TestProgramSwitchesDatabaseViewsWithShiftKeys(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("KITTY_WINDOW_ID", "")
	static := postgresanalysis.BuildGraph(
		[]postgresanalysis.Statement{{QueryID: 42, Query: "SELECT * FROM users"}},
		[]postgresanalysis.Table{{Name: "users"}},
	)
	target := &engine.Engine{
		Service: config.Service{Name: "database", Type: config.ServiceTypeDB, Provider: config.ServiceProviderPostgres},
		Graph:   domain.BuildRuntimeGraph(postgresanalysis.ModulePath, static),
	}
	tm := teatest.NewTestModel(t, initialModel([]*engine.Engine{target}, []config.Service{target.Service}, "", nil),
		teatest.WithInitialTermSize(60, 20),
	)
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("SELECT * FROM users"))
	})
	tm.Type("L")
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("users"))
	})
	tm.Type("H")
	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("SELECT * FROM users"))
	})
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
			1:  {ID: 1, Name: "main", Syntax: domain.Syntax{File: 1}, Out: []domain.NodeID{2, 99}},
			2:  {ID: 2, Name: "handler", In: []domain.NodeID{1}},
			99: {ID: 99, Name: "Serve", Pkg: 9, In: []domain.NodeID{1}},
		},
		Files: map[domain.FileID]*domain.StaticFile{
			1: {ID: 1, Path: "/project/main.go", Functions: []domain.NodeID{1, 2}},
		},
		Packages: map[domain.PackageID]*domain.Package{
			9: {Name: "http", Path: "net/http"},
		},
	}
	return &domain.RuntimeGraph{
		Static: static,
		Nodes: map[domain.NodeID]*domain.Node{
			1: {Static: static.Nodes[1]},
			2: {Static: static.Nodes[2]},
		},
		Files: map[domain.FileID]*domain.File{
			1: {Static: static.Files[1]},
		},
	}
}
