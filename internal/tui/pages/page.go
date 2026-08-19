// Package pages defines the contract shared by TUI pages.
package pages

import (
	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/config"
)

// Page is a selectable service view.
type Page interface {
	Name() string
	Type() config.ServiceType
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() tea.View
}
