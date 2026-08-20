// Package metrics dispatches database dashboards by service provider.
package metrics

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/model"
	postgresmetrics "github.com/briheet/sen/internal/tui/pages/db/metrics/postgres"
)

const (
	screenPercent = 80
	minimumWidth  = 44
	minimumHeight = 10
)

// Model contains exactly one provider-specific dashboard.
type Model struct {
	provider config.ServiceProvider
	postgres postgresmetrics.Model
}

// New selects the dashboard for provider.
func New(provider config.ServiceProvider, source *model.RuntimeGraph) Model {
	result := Model{provider: provider}
	if provider == config.ServiceProviderPostgres {
		result.postgres = postgresmetrics.New(source)
	}
	return result
}

// ApplySnapshot forwards a completed runtime window to the active dashboard.
func (m *Model) ApplySnapshot(metrics model.RuntimeMetrics, sampledAt time.Time) {
	if m.provider == config.ServiceProviderPostgres {
		m.postgres.ApplySnapshot(metrics, sampledAt)
	}
}

// Update forwards terminal state to the active dashboard.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.provider == config.ServiceProviderPostgres {
		m.postgres, _ = m.postgres.Update(msg)
	}
	return m, nil
}

// View renders the selected provider dashboard.
func (m Model) View() string {
	if m.provider == config.ServiceProviderPostgres {
		return m.postgres.View()
	}
	return ""
}

// Size returns a centered modal occupying 80 percent of the viewport.
func Size(width, height int) (int, int) {
	return min(width, max(minimumWidth, (width*screenPercent+50)/100)),
		min(height, max(minimumHeight, (height*screenPercent+50)/100))
}
