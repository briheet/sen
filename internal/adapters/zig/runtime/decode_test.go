package runtime

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const sampleData = "#base 1000\n#interval 1000000\n" +
	"10046edc0 10046ede8 10046edf0\n" +
	"10046edc0 10046ede8\n" +
	"10046edd0\n"

func TestDecodeSamples(t *testing.T) {
	frameOf := func(pc uint64) (string, uint64, bool) {
		return "/app/main.zig", uint64(pc%10) + 1, true
	}

	profile, trace := decodeSamples([]byte(sampleData), time.Millisecond, frameOf)
	require.Len(t, profile.Samples, 3)
	require.Equal(t, 3*time.Millisecond, profile.Duration)
	require.Len(t, profile.Locations, 4)
	require.Len(t, trace.Events, 3)
	require.Equal(t, time.Millisecond, trace.Events[1].At)

	first := profile.Samples[0]
	require.Len(t, first.Stack, 3)
	require.Equal(t, "/app/main.zig", profile.Locations[first.Stack[0]].Frames[0].File)
	require.Equal(t, int64(1), profile.Locations[first.Stack[0]].Frames[0].Line)
}

func TestDecodeSamplesSkipsUnmapped(t *testing.T) {
	frameOf := func(uint64) (string, uint64, bool) { return "", 0, false }

	profile, trace := decodeSamples([]byte("10046edc0 10046ede8\n"), time.Millisecond, frameOf)
	require.Len(t, profile.Samples, 1)
	require.Len(t, profile.Locations, 2)
	require.Empty(t, profile.Locations[1].Frames)
	require.Empty(t, trace.Stacks[1].Frames)
}

func TestParseHeader(t *testing.T) {
	base, interval := parseHeader([]byte(sampleData))
	require.Equal(t, uint64(0x1000), base)
	require.Equal(t, time.Millisecond, interval)
}

func benchSampleData(count int) []byte {
	var builder strings.Builder
	builder.WriteString("#base 0x1000\n#interval 1000000\n")
	for index := 0; index < count; index++ {
		builder.WriteString(" 10046edc0 10046ede8 10046edf0 10046edf4\n")
	}
	return []byte(builder.String())
}

func BenchmarkDecodeSamples(b *testing.B) {
	data := benchSampleData(1000)
	frameOf := func(pc uint64) (string, uint64, bool) {
		return "/app/main.zig", uint64(pc%10) + 1, true
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, trace := decodeSamples(data, time.Millisecond, frameOf); len(trace.Events) == 0 {
			b.Fatal("no events")
		}
	}
}
