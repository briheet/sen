// Package runtime collects metrics and per-query/per-table heat from a running
// PostgreSQL server over its built-in statistics views.
package runtime

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/postgres/analysis"
	"github.com/briheet/sen/internal/adapters/postgres/runtime/metrics"
	"github.com/briheet/sen/internal/adapters/postgres/runtime/trace"
	"github.com/briheet/sen/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const collectionInterval = time.Second

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

	lastCollection  time.Time
	activity        trace.Snapshot
	statementLabels map[int64]string

	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

var _ adapters.Runtime = (*Collector)(nil)

// NewCollector builds a collector that dials the given connection string.
func NewCollector(dsn string) *Collector {
	return &Collector{
		dsn: dsn, done: make(chan struct{}),
		statementLabels: make(map[int64]string),
	}
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
	if err := c.waitForWindow(ctx); err != nil {
		return model.Observation{}, err
	}
	db, err := c.databaseMetrics(ctx)
	if err != nil {
		return model.Observation{}, err
	}
	stmtRows, statementsAvailable, err := c.statementRows(ctx)
	if err != nil {
		return model.Observation{}, err
	}
	tableRows, err := c.tableRows(ctx)
	if err != nil {
		return model.Observation{}, err
	}

	for _, statement := range stmtRows {
		db.StatementCalls += statement.Calls
	}
	db.StatementsAvailable = statementsAvailable

	collectedAt := time.Now()
	duration := collectionInterval
	if !c.lastCollection.IsZero() {
		duration = collectedAt.Sub(c.lastCollection)
	}
	current := trace.NewSnapshot(stmtRows, tableRows)
	var profiles map[string]*model.Profile
	if c.activity.Initialized() {
		profiles = current.Delta(c.activity).Profiles(duration)
	}
	c.activity = current
	c.lastCollection = collectedAt

	return model.Observation{Metrics: metrics.Decode(db), Profiles: profiles}, nil
}

func (c *Collector) waitForWindow(ctx context.Context) error {
	if c.lastCollection.IsZero() {
		return nil
	}
	delay := time.Until(c.lastCollection.Add(collectionInterval))
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Wait blocks for the lifetime of the external PostgreSQL service.
func (c *Collector) Wait() error {
	<-c.done
	return nil
}

// Stop ends observation and releases the connection.
func (c *Collector) Stop() error { return c.close() }

// Cleanup is idempotent and releases the connection.
func (c *Collector) Cleanup() error { return c.close() }

func (c *Collector) close() error {
	c.closeOnce.Do(func() {
		if c.conn != nil {
			c.closeErr = c.conn.Close(context.Background())
		}
		close(c.done)
	})
	return c.closeErr
}

func (c *Collector) databaseMetrics(ctx context.Context) (metrics.Database, error) {
	row := c.conn.QueryRow(ctx, `
		SELECT current_setting('server_version'), current_database(),
		       EXTRACT(EPOCH FROM clock_timestamp() - pg_postmaster_start_time()),
		       pg_database_size(current_database()), current_setting('max_connections')::bigint,
		       d.numbackends,
		       (SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND state = 'active'),
		       (SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND state LIKE 'idle%'),
		       (SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type IS NOT NULL),
		       (SELECT count(*) FROM pg_locks l JOIN pg_database db ON db.oid = l.database WHERE db.datname = current_database()),
		       d.xact_commit, d.xact_rollback, d.blks_read, d.blks_hit,
		       d.tup_returned, d.tup_fetched, d.tup_inserted, d.tup_updated, d.tup_deleted,
		       d.temp_files, d.temp_bytes, d.deadlocks
		FROM pg_stat_database d
		WHERE d.datname = current_database()`)
	return metrics.DatabaseRow(row.Scan)
}

func (c *Collector) statementRows(ctx context.Context) ([]trace.Statement, bool, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT queryid, query, calls, total_exec_time, rows, shared_blks_read
		FROM pg_stat_statements
		WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
		  AND queryid IS NOT NULL AND calls > 0
		ORDER BY total_exec_time DESC`)
	if err != nil {
		if isUndefinedTable(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer rows.Close()

	var out []trace.Statement
	for rows.Next() {
		var statement trace.Statement
		if err := rows.Scan(
			&statement.QueryID, &statement.Query, &statement.Calls,
			&statement.TotalExec, &statement.Rows, &statement.BlocksRead,
		); err != nil {
			return nil, false, err
		}
		statement.Label = c.statementLabels[statement.QueryID]
		if statement.Label == "" {
			statement.Label = analysis.StatementLabel(statement.Query)
			c.statementLabels[statement.QueryID] = statement.Label
		}
		out = append(out, statement)
	}
	return out, true, rows.Err()
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
