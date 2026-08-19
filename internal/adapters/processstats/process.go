// Package processstats collects operating-system metrics for a target process.
package processstats

import (
	"context"
	"runtime"
	"time"

	"github.com/briheet/sen/internal/model"
	"github.com/shirou/gopsutil/v4/process"
)

// Sampler reads one target process and retains its peak resident memory.
type Sampler struct {
	process *process.Process
	peakRSS uint64
}

// New opens a process sampler for pid.
func New(ctx context.Context, pid int) (*Sampler, error) {
	target, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return nil, err
	}
	return &Sampler{process: target}, nil
}

// Collect returns every measurement available on the current platform.
func (s *Sampler) Collect(ctx context.Context) model.ProcessMetrics {
	var result model.ProcessMetrics
	if times, err := s.process.TimesWithContext(ctx); err == nil {
		result.UserCPU = times.User
		result.SystemCPU = times.System
		result.Available |= model.ProcessCPU
	}
	if memory, err := s.process.MemoryInfoWithContext(ctx); err == nil {
		result.RSS = memory.RSS
		s.peakRSS = max(s.peakRSS, memory.RSS)
		result.PeakRSS = s.peakRSS
		result.VirtualMemory = memory.VMS
		result.Available |= model.ProcessMemory
	}
	if threads, err := s.process.NumThreadsWithContext(ctx); err == nil && threads >= 0 {
		result.Threads = uint64(threads)
		result.Available |= model.ProcessThreads
	}
	if files, err := s.process.NumFDsWithContext(ctx); err == nil && files >= 0 {
		result.OpenFiles = uint64(files)
		result.Available |= model.ProcessOpenFiles
	}
	if counters, err := s.process.IOCountersWithContext(ctx); err == nil {
		result.ReadBytes = max(counters.ReadBytes, counters.DiskReadBytes)
		result.WriteBytes = max(counters.WriteBytes, counters.DiskWriteBytes)
		result.Available |= model.ProcessIO
		// Darwin exposes disk bytes but not operation counts.
		if runtime.GOOS != "darwin" {
			result.ReadOps = counters.ReadCount
			result.WriteOps = counters.WriteCount
			result.Available |= model.ProcessIOOperations
		}
	}
	if voluntary, involuntary, ok := contextSwitches(ctx, s.process); ok {
		result.VoluntaryCS = voluntary
		result.InvoluntaryCS = involuntary
		result.Available |= model.ProcessContextSwitches
	}
	if started, err := s.process.CreateTimeWithContext(ctx); err == nil {
		result.Uptime = max(0, time.Since(time.UnixMilli(started)))
		result.Available |= model.ProcessUptime
	}
	return result
}
