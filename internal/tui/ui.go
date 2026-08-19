// This file mainly deals with ui decisions of the tui
package tui

import (
	"io"

	tea "charm.land/bubbletea/v2"

	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/components/footer"
	"github.com/briheet/sen/internal/tui/components/header"
	"github.com/briheet/sen/internal/tui/components/keys"
	tuicontext "github.com/briheet/sen/internal/tui/context"
	"github.com/briheet/sen/internal/tui/pages"
	"github.com/briheet/sen/internal/tui/pages/kv"
	"github.com/briheet/sen/internal/tui/pages/servers"
	"github.com/briheet/sen/internal/tui/styles"
)

// enginesDoneMsg closes the TUI after every configured process exits.
type enginesDoneMsg struct{}

// TODO: After Redis gets added, remove services passed in this
// and use the engine config to switch on types.
func initialModel(engines []*engine.Engine, services []config.Service, logPath string, dump io.Writer) model {
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
			page = servers.New(target, dump)
		case config.ServiceTypeKV:
			if target == nil {
				page = kv.FromService(service)
			} else {
				page = kv.New(target)
			}
		default:
			continue
		}
		pageModels = append(pageModels, page)
	}

	ctx := tuicontext.New(engines, pageModels, logPath)
	theme := styles.Zakura
	keyMap := keys.NewModel()

	return model{
		dump:   dump,
		ctx:    ctx,
		header: header.NewModel(ctx),
		footer: footer.NewModel(keyMap, theme),
		keys:   keyMap,

		activeTheme: theme,
	}
}

// Init starts every service page.
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.header.Init(), m.footer.Init(), m.initScreen())
}

// initScreen initializes every page concurrently.
func (m model) initScreen() tea.Cmd {
	pages := m.ctx.Pages()
	commands := make([]tea.Cmd, 0, len(pages))
	for _, page := range pages {
		commands = append(commands, page.Init())
	}
	return tea.Batch(commands...)
}
