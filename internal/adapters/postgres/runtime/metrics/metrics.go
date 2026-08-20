// Package metrics maps PostgreSQL cumulative statistics into normalized runtime
// metrics. Only the built-in statistics views are used, so no extension or
// special configuration is required beyond a read-only connection.
package metrics

import (
	"time"

	"github.com/briheet/sen/internal/model"
)

// Database is one row of pg_stat_database folded into RuntimeMetrics.
type Database struct {
	Version             string
	Name                string
	Uptime              time.Duration
	Size                int64
	MaxConnections      int64
	Backends            int64
	Active              int64
	Idle                int64
	Waiting             int64
	Locks               int64
	Commits             int64
	Rollbacks           int64
	BlocksRead          int64
	BlocksHit           int64
	TuplesReturned      int64
	TuplesFetched       int64
	TuplesIn            int64
	TuplesUpd           int64
	TuplesDel           int64
	TempFiles           int64
	TempBytes           int64
	Deadlocks           int64
	StatementCalls      int64
	StatementsAvailable bool
}

// DatabaseRow scans a pg_stat_database row into a Database.
func DatabaseRow(scan func(dst ...any) error) (Database, error) {
	var d Database
	var uptimeSeconds float64
	err := scan(
		&d.Version, &d.Name, &uptimeSeconds, &d.Size, &d.MaxConnections,
		&d.Backends, &d.Active, &d.Idle, &d.Waiting, &d.Locks,
		&d.Commits, &d.Rollbacks, &d.BlocksRead, &d.BlocksHit,
		&d.TuplesReturned, &d.TuplesFetched, &d.TuplesIn, &d.TuplesUpd, &d.TuplesDel,
		&d.TempFiles, &d.TempBytes, &d.Deadlocks,
	)
	d.Uptime = time.Duration(max(0, uptimeSeconds) * float64(time.Second))
	return d, err
}

// Decode folds a database statistics row into PostgreSQL runtime metrics.
func Decode(db Database) *model.RuntimeMetrics {
	return &model.RuntimeMetrics{Postgres: model.PostgresMetrics{
		Version:             db.Version,
		Database:            db.Name,
		Uptime:              db.Uptime,
		DatabaseSize:        nonNegative(db.Size),
		MaxConnections:      nonNegative(db.MaxConnections),
		Backends:            nonNegative(db.Backends),
		Active:              nonNegative(db.Active),
		Idle:                nonNegative(db.Idle),
		Waiting:             nonNegative(db.Waiting),
		Locks:               nonNegative(db.Locks),
		Commits:             nonNegative(db.Commits),
		Rollbacks:           nonNegative(db.Rollbacks),
		BlocksRead:          nonNegative(db.BlocksRead),
		BlocksHit:           nonNegative(db.BlocksHit),
		TuplesReturned:      nonNegative(db.TuplesReturned),
		TuplesFetched:       nonNegative(db.TuplesFetched),
		TuplesIn:            nonNegative(db.TuplesIn),
		TuplesUpd:           nonNegative(db.TuplesUpd),
		TuplesDel:           nonNegative(db.TuplesDel),
		TempFiles:           nonNegative(db.TempFiles),
		TempBytes:           nonNegative(db.TempBytes),
		Deadlocks:           nonNegative(db.Deadlocks),
		StatementCalls:      nonNegative(db.StatementCalls),
		StatementsAvailable: db.StatementsAvailable,
	}}
}

func nonNegative(value int64) uint64 { return uint64(max(0, value)) }
