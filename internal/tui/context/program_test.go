package context

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/stretchr/testify/require"
)

type testPage struct{ name string }

func (p testPage) Name() string                         { return p.name }
func (testPage) Type() config.ServiceType               { return config.ServiceTypeServer }
func (testPage) Init() tea.Cmd                          { return nil }
func (p testPage) Update(tea.Msg) (pages.Page, tea.Cmd) { return p, nil }
func (p testPage) View() tea.View                       { return tea.NewView(p.name) }
func (testPage) Revision() uint64                       { return 0 }

func TestPagesReusesOrderedView(t *testing.T) {
	ctx := New(nil, []pages.Page{testPage{name: "api"}, testPage{name: "worker"}}, "")

	require.Zero(t, testing.AllocsPerRun(100, func() { _ = ctx.Pages() }))
	ctx.SetPage(testPage{name: "api"})
	require.Equal(t, []string{"api", "worker"}, []string{ctx.Pages()[0].Name(), ctx.Pages()[1].Name()})
}
