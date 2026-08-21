package pages

import (
	"time"
)

// TelemetryTickMsg delivers a completed engine snapshot to its service page.
type TelemetryTickMsg struct {
	At      time.Time
	Service string
}
