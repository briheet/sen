package cpuprofile

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/briheet/senbon/internal/model"
	"github.com/stretchr/testify/require"
)

const fixture = `{
	"nodes": [
		{"id":1,"callFrame":{"functionName":"(root)","lineNumber":-1,"columnNumber":-1,"url":""},"children":[2]},
		{"id":2,"callFrame":{"functionName":"main","lineNumber":0,"columnNumber":0,"url":"file:///app/index.js"},"children":[3]},
		{"id":3,"callFrame":{"functionName":"work","lineNumber":5,"columnNumber":2,"url":"/app/index.js"}}
	],
	"startTime":1000,
	"endTime":6000,
	"samples":[3,3,2,3,2],
	"timeDeltas":[1000,1000,1000,1000,1000]
}`

func fixtureCPUProfile(t *testing.T) *CPUProfile {
	t.Helper()
	var profile CPUProfile
	require.NoError(t, json.Unmarshal([]byte(fixture), &profile))
	return &profile
}

func TestProfile(t *testing.T) {
	profile := fixtureCPUProfile(t)
	result := profile.Profile()
	require.Equal(t, 5*time.Millisecond, result.Duration)
	require.Equal(t, []model.ValueType{{Type: "cpu", Unit: "nanoseconds"}}, result.SampleTypes)
	require.Len(t, result.Locations, 3)
	require.Len(t, result.Samples, 5)

	work := result.Locations[3]
	require.Equal(t, "work", work.Frames[0].Function)
	require.Equal(t, "/app/index.js", work.Frames[0].File)
	require.Equal(t, int64(6), work.Frames[0].Line)

	first := result.Samples[0]
	require.Equal(t, []int64{1000000}, first.Values)
	require.Equal(t, []model.ProfileLocationID{3, 2, 1}, first.Stack)

	second := result.Samples[2]
	require.Equal(t, []model.ProfileLocationID{2, 1}, second.Stack)
}

func TestTrace(t *testing.T) {
	profile := fixtureCPUProfile(t)
	trace := profile.Trace()
	require.Equal(t, 5*time.Millisecond, trace.Duration)
	require.Len(t, trace.Events, 5)
	require.Len(t, trace.Stacks, 2)

	require.Equal(t, time.Millisecond, trace.Events[0].At)
	require.Equal(t, model.EventStackSample, trace.Events[0].Kind)
	require.Equal(t, 5*time.Millisecond, trace.Events[4].At)

	for index, event := range trace.Events {
		require.NotZero(t, event.Stack)
		frames := trace.Stacks[event.Stack].Frames
		if index == 2 || index == 4 {
			require.Len(t, frames, 2)
			require.Equal(t, "main", frames[0].Function)
			require.Equal(t, "(root)", frames[1].Function)
			continue
		}
		require.Len(t, frames, 3)
		require.Equal(t, "work", frames[0].Function)
		require.Equal(t, uint64(6), frames[0].Line)
		require.Equal(t, "main", frames[1].Function)
		require.Equal(t, "/app/index.js", frames[1].File)
	}
}
