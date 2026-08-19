//go:build linux

package processstats

import (
	"context"

	"github.com/shirou/gopsutil/v4/process"
)

func contextSwitches(ctx context.Context, target *process.Process) (uint64, uint64, bool) {
	result, err := target.NumCtxSwitchesWithContext(ctx)
	if err != nil || result.Voluntary < 0 || result.Involuntary < 0 {
		return 0, 0, false
	}
	return uint64(result.Voluntary), uint64(result.Involuntary), true
}
