// Package runtime collects metrics and per-query/per-table heat from a running
// PostgreSQL server over its built-in statistics views.
package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/postgres/runtime/metrics"
	"github.com/briheet/sen/internal/adapters/postgres/runtime/trace"
	"github.com/briheet/sen/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// isUndefinedTable reports whether err is PostgreSQL error 42P01, the
// undefined_table error raised when pg_stat_statements is not installed.
func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

// Collector owns a connection to a running PostgreSQL server.
type Collector struct {
	dsn  string
	conn *pgx.Conn

	Metrics    *model.RuntimeMetrics
	Statements *model.Profile
	Tables     *model.Profile

	doneOnce sync.Once
	done     chan struct{}
}

var _ adapters.Runtime = (*Collector)(nil)

// NewCollector builds a collector that dials the given connection string.
func NewCollector(dsn string) *Collector {
	return &Collector{dsn: dsn, done: make(chan struct{})}
}

// Start connects and verifies the server is reachable.
func (c *Collector) Start(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, c.dsn)
	if err != nil {
		return err
	}
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close(ctx)
		return err
	}
	c.conn = conn
	return nil
}

// Collect pulls one snapshot: server metrics plus per-query and per-table heat.
func (c *Collector) Collect(ctx context.Context) (model.Observation, error) {
	db, err := c.databaseMetrics(ctx)
	if err != nil {
		return model.Observation{}, err
	}
	stmtRows, err := c.statementRows(ctx)
	if err != nil {
		return model.Observation{}, err
	}
	tableRows, err := c.tableRows(ctx)
	if err != nil {
		return model.Observation{}, err
	}

	c.Metrics = metrics.Decode(db)
	c.Statements = trace.Statements(stmtRows)
	c.Tables = trace.Tables(tableRows)

	profiles := map[string]*model.Profile{}
	if len(c.Statements.Samples) > 0 {
		profiles[trace.StatementsSource] = c.Statements
	}
	if len(c.Tables.Samples) > 0 {
		profiles[trace.TablesSource] = c.Tables
	}
	c.finish()
	return model.Observation{Metrics: c.Metrics, Profiles: profiles}, nil
}

// Wait blocks until the first snapshot is collected or Stop is called.
func (c *Collector) Wait() error {
	<-c.done
	return nil
}

// Stop ends observation and releases the connection.
func (c *Collector) Stop() error {
	c.finish()
	return c.close()
}

// Cleanup releases the connection.
func (c *Collector) Cleanup() error {
	c.finish()
	return c.close()
}

func (c *Collector) close() error {
	if c.conn != nil {
		return c.conn.Close(context.Background())
	}
	return nil
}

// finish unblocks any waiter exactly once.
func (c *Collector) finish() {
	c.doneOnce.Do(func() { close(c.done) })
}

func (c *Collector) databaseMetrics(ctx context.Context) (metrics.Database, error) {
	row := c.conn.QueryRow(ctx, `
		SELECT numbackends, xact_commit, xact_rollback,
		       blks_read, blks_hit,
		       tup_inserted, tup_updated, tup_deleted,
		       temp_files, deadlocks
		FROM pg_stat_database
		WHERE datname = current_database()`)
	return metrics.DatabaseRow(row.Scan)
}

func (c *Collector) statementRows(ctx context.Context) ([]trace.Statement, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT queryid, query, calls, total_exec_time, rows, shared_blks_read
		FROM pg_stat_statements
		WHERE queryid IS NOT NULL AND calls > 0
		ORDER BY total_exec_time DESC`)
	if err != nil {
		if isUndefinedTable(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var out []trace.Statement
	for rows.Next() {
		var s trace.Statement
		if err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalExec, &s.Rows, &s.BlocksRead); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (c *Collector) tableRows(ctx context.Context) ([]trace.Table, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT relname, seq_scan, idx_scan,
		       n_tup_ins, n_tup_upd, n_tup_del, n_live_tup
		FROM pg_stat_user_tables
		ORDER BY relname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []trace.Table
	for rows.Next() {
		var t trace.Table
		if err := rows.Scan(&t.Name, &t.SeqScan, &t.IdxScan, &t.Inserts, &t.Updates, &t.Deletes, &t.LiveTuples); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
