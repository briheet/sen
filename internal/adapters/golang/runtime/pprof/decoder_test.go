package pprof

import (
	"bytes"
	runtimepprof "runtime/pprof"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadRuntimeProfile(t *testing.T) {
	var data bytes.Buffer
	require.NoError(t, runtimepprof.Lookup("goroutine").WriteTo(&data, 0))

	result, err := Read(bytes.NewReader(data.Bytes()))
	require.NoError(t, err)
	require.NotEmpty(t, result.SampleTypes)
	require.NotEmpty(t, result.Samples)
	require.NotEmpty(t, result.Locations)

	var found bool
	for _, location := range result.Locations {
		for _, frame := range location.Frames {
			found = found || frame.Function == "github.com/briheet/sen/internal/adapters/golang/runtime/pprof.TestReadRuntimeProfile"
		}
	}
	require.True(t, found)
}
