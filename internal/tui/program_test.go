package tui

import (
	"bytes"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/stretchr/testify/require"
)

func TestProgramNavigatesPages(t *testing.T) {
	engines := []*engine.Engine{
		{Service: config.Service{Name: "api", Type: config.ServiceTypeServer, Lang: config.ServiceLangGo}},
		{Service: config.Service{Name: "worker", Type: config.ServiceTypeServer, Lang: config.ServiceLangNode}},
	}
	services := []config.Service{
		engines[0].Service,
		{Name: "cache", Type: config.ServiceTypeKV, Provider: config.ServiceProviderRedis, Address: "localhost:6379"},
		engines[1].Service,
	}
	tm := teatest.NewTestModel(t, initialModel(engines, services, "/tmp/engine.log"),
		teatest.WithInitialTermSize(80, 24),
	)
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
		return bytes.Contains(output, []byte("api (go)")) &&
			bytes.Contains(output, []byte("more"))
	})
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
	tm := teatest.NewTestModel(t, initialModel(nil, nil, ""))
	t.Cleanup(func() { require.NoError(t, tm.Quit()) })

	tm.Send(enginesDoneMsg{})
	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))

	require.IsType(t, model{}, final)
}
