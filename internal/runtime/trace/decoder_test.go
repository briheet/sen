package trace

import (
	"bytes"
	"context"
	"runtime/trace"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadRuntimeTrace(t *testing.T) {
	var data bytes.Buffer
	require.NoError(t, trace.Start(&data))

	ctx, task := trace.NewTask(context.Background(), "build")
	trace.WithRegion(ctx, "compile", func() {
		trace.Log(ctx, "package", "example.com/app")
	})
	task.End()
	trace.Stop()

	result, err := Read(bytes.NewReader(data.Bytes()))
	require.NoError(t, err)
	require.NotEmpty(t, result.Events)
	require.NotEmpty(t, result.Stacks)

	var taskBegin, regionBegin, logEvent bool
	for _, event := range result.Events {
		switch event.Kind {
		case EventTaskBegin:
			taskBegin = taskBegin || event.Name == "build"
		case EventRegionBegin:
			regionBegin = regionBegin || event.Name == "compile"
		case EventLog:
			logEvent = logEvent || event.Category == "package" && event.Message == "example.com/app"
		}
	}
	require.True(t, taskBegin)
	require.True(t, regionBegin)
	require.True(t, logEvent)
}
