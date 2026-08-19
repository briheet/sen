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
	Revision() uint64
}

// ViewportMsg gives a page its absolute terminal bounds and visibility.
type ViewportMsg struct {
	X, Y          int
	Width, Height int
	Visible       bool
}

// ObscuredMsg softens page visuals while a modal is shown above them.
type ObscuredMsg struct{ Obscured bool }
