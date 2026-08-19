//go:build !linux

package processstats

import (
	"context"

	"github.com/shirou/gopsutil/v4/process"
)

func contextSwitches(context.Context, *process.Process) (uint64, uint64, bool) {
	return 0, 0, false
}
