package pages

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const telemetryInterval = 250 * time.Millisecond

// TelemetryTickMsg lets the active page consume completed engine snapshots.
type TelemetryTickMsg struct{ At time.Time }

// NextTelemetryTick keeps collection updates on Bubble Tea's event loop.
func NextTelemetryTick() tea.Cmd {
	return tea.Tick(telemetryInterval, func(at time.Time) tea.Msg { return TelemetryTickMsg{At: at} })
}
