package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	m := Decode(Database{
		Backends:   4,
		Commits:    1000,
		Rollbacks:  10,
		BlocksRead: 500,
		BlocksHit:  50000,
		TuplesIn:   200,
		TuplesUpd:  50,
		TuplesDel:  5,
		TempFiles:  3,
		Deadlocks:  0,
	})

	assert.Equal(t, uint64(4), m.Postgres.Backends)
	assert.Equal(t, uint64(1000), m.Postgres.Commits)
	assert.Equal(t, uint64(10), m.Postgres.Rollbacks)
	assert.Equal(t, uint64(500), m.Postgres.BlocksRead)
	assert.Equal(t, uint64(50000), m.Postgres.BlocksHit)
	assert.Equal(t, uint64(200), m.Postgres.TuplesIn)
	assert.Equal(t, uint64(50), m.Postgres.TuplesUpd)
	assert.Equal(t, uint64(5), m.Postgres.TuplesDel)
}

func TestDecodeZero(t *testing.T) {
	t.Parallel()

	m := Decode(Database{})
	assert.Equal(t, uint64(0), m.Postgres.Backends)
	assert.Equal(t, uint64(0), m.Postgres.Commits)
}
