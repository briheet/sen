package context

import (
	"github.com/briheet/sen/internal/engine"
	"github.com/briheet/sen/internal/tui/pages"
)

// ProgramContext contains state shared by workspace components.
type ProgramContext struct {
	// Collection of all engines from all adapters.
	engines []*engine.Engine
	// TODO: Currently services traces and application both
	// log in a single file, please change this behaviour.
	logPath string // engine logging path for UI logs

	pages      map[string]pages.Page // service pages by name
	pageOrder  []string              // service names in configuration order
	activePage string                // current active page
}

// New creates the shared state used by the workspace and its components.
func New(engines []*engine.Engine, pageModels []pages.Page, logPath string) *ProgramContext {
	c := &ProgramContext{
		engines:   engines,
		logPath:   logPath,
		pages:     make(map[string]pages.Page, len(pageModels)),
		pageOrder: make([]string, 0, len(pageModels)),
	}
	for _, page := range pageModels {
		c.pages[page.Name()] = page
		c.pageOrder = append(c.pageOrder, page.Name())
		if c.activePage == "" {
			c.activePage = page.Name()
		}
	}
	return c
}

// Pages returns service pages in configuration order.
func (c *ProgramContext) Pages() []pages.Page {
	result := make([]pages.Page, 0, len(c.pageOrder))
	for _, name := range c.pageOrder {
		result = append(result, c.pages[name])
	}
	return result
}

// Page returns a service page by name.
func (c *ProgramContext) Page(name string) (pages.Page, bool) {
	page, ok := c.pages[name]
	return page, ok
}

// SetPage stores state returned by a page update.
func (c *ProgramContext) SetPage(page pages.Page) {
	c.pages[page.Name()] = page
}

// ActivePage returns the selected service name.
func (c *ProgramContext) ActivePage() string {
	return c.activePage
}

// SelectPage selects an existing service page.
func (c *ProgramContext) SelectPage(name string) {
	if _, ok := c.pages[name]; ok {
		c.activePage = name
	}
}

// LogPath returns the engine log path for this run.
func (c *ProgramContext) LogPath() string {
	return c.logPath
}
