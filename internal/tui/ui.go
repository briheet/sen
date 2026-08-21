// This file mainly deals with ui decisions of the tui
package tui

import (
	"io"

	tea "charm.land/bubbletea/v2"

	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/components/keys"
	"github.com/briheet/sen/internal/tui/components/statusbar"
	tuicontext "github.com/briheet/sen/internal/tui/context"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/pages/db"
	"github.com/briheet/sen/internal/tui/pages/kv"
	"github.com/briheet/sen/internal/tui/pages/servers"
	"github.com/briheet/sen/internal/tui/styles"
)

// enginesDoneMsg closes the TUI after every configured process exits.
type enginesDoneMsg struct{}

func initialModel(engines []*engine.Engine, services []config.Service, logPath string, dump io.Writer, configuredTheme ...styles.Theme) *model {
	theme := styles.Zakura
	if len(configuredTheme) > 0 {
		theme = configuredTheme[0]
	}
	pageModels := make([]pages.Page, 0, len(services))
	targets := make(map[string]*engine.Engine, len(engines))
	for _, target := range engines {
		targets[target.Service.Name] = target
	}

	// Build pages in configuration order.
	for _, service := range services {
		var page pages.Page
		target := targets[service.Name]
		switch service.Type {
		case config.ServiceTypeServer:
			if target == nil {
				continue
			}
			page = servers.NewWithTheme(target, dump, theme)
		case config.ServiceTypeKV:
			if target == nil {
				page = kv.FromServiceWithTheme(service, theme)
			} else {
				page = kv.NewWithTheme(target, dump, theme)
			}
		case config.ServiceTypeDB:
			if target == nil {
				page = db.FromServiceWithTheme(service, theme)
			} else {
				page = db.NewWithTheme(target, dump, theme)
			}
		default:
			continue
		}
		pageModels = append(pageModels, page)
	}

	ctx := tuicontext.New(engines, pageModels, logPath)
	keyMap := keys.NewModel()

	m := &model{
		dump:      dump,
		ctx:       ctx,
		statusbar: statusbar.New(ctx, keyMap, theme),
		keys:      keyMap,

		activeTheme: theme,
	}
	m.refreshView()
	return m
}

// Init starts every service page.
func (m *model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.statusbar.Init(), m.initScreen())
}

// initScreen initializes every page concurrently.
func (m *model) initScreen() tea.Cmd {
	pages := m.ctx.Pages()
	commands := make([]tea.Cmd, 0, len(pages))
	for _, page := range pages {
		commands = append(commands, page.Init())
	}
	return tea.Batch(commands...)
}
