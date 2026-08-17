package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatements(t *testing.T) {
	t.Parallel()

	p := Statements([]Statement{
		{QueryID: 111, Query: "SELECT * FROM users WHERE id = $1", Calls: 5, TotalExec: 2.5},
		{QueryID: 222, Query: "INSERT INTO orders", Calls: 3, TotalExec: 1.0},
	})
	require.NotNil(t, p)
	assert.Len(t, p.Samples, 2)
	assert.Len(t, p.Locations, 2)

	saw := map[int64]bool{}
	for _, s := range p.Samples {
		require.Len(t, s.Stack, 1)
		saw[s.Values[0]] = true
	}
	assert.True(t, saw[5])
	assert.True(t, saw[3])
}

func TestStatementsPath(t *testing.T) {
	t.Parallel()

	p := Statements([]Statement{{QueryID: 42, Query: "SELECT 1", Calls: 1, TotalExec: 0.1}})
	loc := p.Locations[p.Samples[0].Stack[0]]
	assert.Equal(t, "sen/postgres/stmt/42", loc.Frames[0].File)
	assert.Equal(t, "SELECT 1", loc.Frames[0].Function)
}

func TestTables(t *testing.T) {
	t.Parallel()

	p := Tables([]Table{{Name: "users", SeqScan: 4, LiveTuples: 200}})
	require.NotNil(t, p)
	require.Len(t, p.Samples, 1)
	loc := p.Locations[p.Samples[0].Stack[0]]
	assert.Equal(t, "sen/postgres/table/users", loc.Frames[0].File)
	assert.Equal(t, int64(4), p.Samples[0].Values[0])
}

func TestLabelTruncation(t *testing.T) {
	t.Parallel()

	long := "SELECT " + repeat("x", 100)
	assert.LessOrEqual(t, len([]rune(label(long))), maxLabel+1) // +1 rune for ellipsis
	assert.Equal(t, "SELECT 1", label("SELECT 1\n"))
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
