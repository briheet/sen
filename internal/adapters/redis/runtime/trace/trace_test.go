package trace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	body := "# Commandstats\r\n" +
		"cmdstat_get:calls=3,usec=9,usec_per_call=3.00,rejected_calls=0,failed_calls=0\r\n" +
		"cmdstat_set:calls=5,usec=50,usec_per_call=10.00,rejected_calls=0,failed_calls=0\r\n" +
		"cmdstat_unknowncmd:calls=1,usec=1,usec_per_call=1.00,rejected_calls=0,failed_calls=0\r\n"

	p := Parse(body).Profile(time.Second)
	require.NotNil(t, p)
	assert.Len(t, p.Samples, 2)
	assert.Len(t, p.Locations, 2)

	got := make(map[string]struct {
		calls, nanoseconds int64
	})
	for _, sample := range p.Samples {
		require.Len(t, sample.Stack, 1)
		loc := p.Locations[sample.Stack[0]]
		require.Len(t, loc.Frames, 1)
		frame := loc.Frames[0]
		got[frame.Function] = struct {
			calls, nanoseconds int64
		}{sample.Values[0], sample.Values[1]}
	}

	assert.Equal(t, int64(3), got["GET"].calls)
	assert.Equal(t, int64(9000), got["GET"].nanoseconds)
	assert.Equal(t, int64(5), got["SET"].calls)
	assert.Equal(t, int64(50000), got["SET"].nanoseconds)
}

func TestSnapshotDelta(t *testing.T) {
	previous := Snapshot{"GET": {Calls: 10, Microseconds: 25}}
	current := Snapshot{
		"GET": {Calls: 14, Microseconds: 40},
		"SET": {Calls: 2, Microseconds: 6},
	}

	delta := current.Delta(previous)
	assert.Equal(t, Counters{Calls: 4, Microseconds: 15}, delta["GET"])
	assert.Equal(t, Counters{Calls: 2, Microseconds: 6}, delta["SET"])
}

func TestDecodeEmpty(t *testing.T) {
	t.Parallel()

	p := Parse("").Profile(time.Second)
	require.NotNil(t, p)
	assert.Empty(t, p.Samples)
}
