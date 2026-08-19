package processstats

import (
	"context"
	"os"
	"testing"

	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/require"
)

func TestCollectCurrentProcess(t *testing.T) {
	sampler, err := New(context.Background(), os.Getpid())
	require.NoError(t, err)

	metrics := sampler.Collect(context.Background())
	require.True(t, metrics.Has(model.ProcessCPU))
	require.True(t, metrics.Has(model.ProcessMemory))
	require.True(t, metrics.Has(model.ProcessThreads))
	require.True(t, metrics.Has(model.ProcessOpenFiles))
	require.True(t, metrics.Has(model.ProcessUptime))
	require.Positive(t, metrics.RSS)
	require.GreaterOrEqual(t, metrics.PeakRSS, metrics.RSS)
	require.Positive(t, metrics.Threads)
}
