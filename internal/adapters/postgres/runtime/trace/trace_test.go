package trace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatements(t *testing.T) {
	t.Parallel()

	p := NewSnapshot([]Statement{
		{QueryID: 111, Query: "SELECT * FROM users WHERE id = $1", Calls: 5, TotalExec: 2.5},
		{QueryID: 222, Query: "INSERT INTO orders", Calls: 3, TotalExec: 1.0},
	}, nil).statementProfile(time.Second)
	require.NotNil(t, p)
	assert.Len(t, p.Samples, 2)
	assert.Len(t, p.Locations, 2)

	saw := map[int64]bool{}
	for _, sample := range p.Samples {
		require.Len(t, sample.Stack, 1)
		saw[sample.Values[0]] = true
	}
	assert.True(t, saw[5])
	assert.True(t, saw[3])
}

func TestStatementsPath(t *testing.T) {
	t.Parallel()

	p := NewSnapshot([]Statement{{QueryID: 42, Query: "SELECT 1", Calls: 1, TotalExec: 0.1}}, nil).
		statementProfile(time.Second)
	location := p.Locations[p.Samples[0].Stack[0]]
	assert.Equal(t, "sen/postgres/stmt/42", location.Frames[0].File)
	assert.Equal(t, "SELECT 1", location.Frames[0].Function)
}

func TestTables(t *testing.T) {
	t.Parallel()

	p := NewSnapshot(nil, []Table{{Name: "users", SeqScan: 4, LiveTuples: 200}}).tableProfile(time.Second)
	require.NotNil(t, p)
	require.Len(t, p.Samples, 1)
	loc := p.Locations[p.Samples[0].Stack[0]]
	assert.Equal(t, "sen/postgres/table/users", loc.Frames[0].File)
	assert.Equal(t, int64(4), p.Samples[0].Values[0])
}

func TestSnapshotDeltaUsesOnlyCurrentWindow(t *testing.T) {
	previous := NewSnapshot(
		[]Statement{{QueryID: 42, Query: "SELECT 1", Calls: 10, TotalExec: 5}},
		[]Table{{Name: "users", SeqScan: 4, Updates: 2}},
	)
	current := NewSnapshot(
		[]Statement{{QueryID: 42, Query: "SELECT 1", Calls: 13, TotalExec: 6.5}},
		[]Table{{Name: "users", SeqScan: 5, Updates: 4}},
	)
	profiles := current.Delta(previous).Profiles(time.Second)

	require.Equal(t, int64(3), profiles[StatementsSource].Samples[0].Values[0])
	require.Equal(t, int64(1_500_000), profiles[StatementsSource].Samples[0].Values[1])
	require.Equal(t, int64(3), profiles[TablesSource].Samples[0].Values[0])
}
