package statusbar

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/context"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/stretchr/testify/require"
)

type testPage struct{ name string }

func (p testPage) Name() string                         { return p.name }
func (testPage) Type() config.ServiceType               { return config.ServiceTypeServer }
func (testPage) Init() tea.Cmd                          { return nil }
func (p testPage) Update(tea.Msg) (pages.Page, tea.Cmd) { return p, nil }
func (testPage) View() tea.View                         { return tea.NewView("") }
func (testPage) Revision() uint64                       { return 0 }

func TestStatusBarNavigatesServicesAndOpensHelp(t *testing.T) {
	ctx := context.New(nil, []pages.Page{testPage{"api"}, testPage{"cache"}}, "")
	model := New(ctx, keys.NewModel(), styles.Zakura)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 60, Height: 1})

	require.Equal(t, 60, lipgloss.Width(model.View()))
	require.Contains(t, model.View(), "? help")
	require.Zero(t, testing.AllocsPerRun(100, func() { _ = model.View() }))

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	require.Equal(t, "cache", ctx.ActivePage())
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	serviceX := lipgloss.Width(model.leftView()) + lipgloss.Width(model.style.Services.Selected.Render("api")) + 1
	model, _ = model.Update(tea.MouseClickMsg{X: serviceX, Y: 0, Button: tea.MouseLeft})
	require.Equal(t, "cache", ctx.ActivePage())
	model, _ = model.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	panel := model.HelpView()
	require.Contains(t, panel, "move graph")
	require.LessOrEqual(t, lipgloss.Width(panel), 60)
}
