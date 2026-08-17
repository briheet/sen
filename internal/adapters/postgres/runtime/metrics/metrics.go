// Package metrics maps PostgreSQL cumulative statistics into normalized runtime
// metrics. Only the built-in statistics views are used, so no extension or
// special configuration is required beyond a read-only connection.
package metrics

import "github.com/briheet/sen/internal/model"

// Database is one row of pg_stat_database folded into RuntimeMetrics.
type Database struct {
	Backends   int64
	Commits    int64
	Rollbacks  int64
	BlocksRead int64
	BlocksHit  int64
	TuplesIn   int64
	TuplesUpd  int64
	TuplesDel  int64
	TempFiles  int64
	Deadlocks  int64
}

// DatabaseRow scans a pg_stat_database row into a Database.
func DatabaseRow(scan func(dst ...any) error) (Database, error) {
	var d Database
	err := scan(&d.Backends, &d.Commits, &d.Rollbacks, &d.BlocksRead, &d.BlocksHit,
		&d.TuplesIn, &d.TuplesUpd, &d.TuplesDel, &d.TempFiles, &d.Deadlocks)
	return d, err
}

// Decode folds a database statistics row into PostgreSQL runtime metrics.
func Decode(db Database) *model.RuntimeMetrics {
	return &model.RuntimeMetrics{Postgres: model.PostgresMetrics{
		Backends:   nonNegative(db.Backends),
		Commits:    nonNegative(db.Commits),
		Rollbacks:  nonNegative(db.Rollbacks),
		BlocksRead: nonNegative(db.BlocksRead),
		BlocksHit:  nonNegative(db.BlocksHit),
		TuplesIn:   nonNegative(db.TuplesIn),
		TuplesUpd:  nonNegative(db.TuplesUpd),
		TuplesDel:  nonNegative(db.TuplesDel),
		TempFiles:  nonNegative(db.TempFiles),
		Deadlocks:  nonNegative(db.Deadlocks),
	}}
}

func nonNegative(value int64) uint64 { return uint64(max(0, value)) }
