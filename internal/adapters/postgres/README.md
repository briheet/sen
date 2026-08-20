# PostgreSQL adapter

The PostgreSQL adapter observes an external database through its statistics
views; it never injects code or modifies application queries.

- `analysis` builds a synthetic graph from `pg_stat_statements` and
  `pg_stat_user_tables`.
- `runtime/metrics` maps `pg_stat_database`, `pg_stat_activity`, and
  `pg_locks` into typed PostgreSQL telemetry.
- `runtime/trace` differences cumulative statement and table counters so each
  graph update represents only the latest collection window.
- `pages/db` projects that shared attribution graph into separate statement
  and table views without collecting the database twice.

`pg_stat_statements` is optional. Without it, database and table metrics still
work, while statement nodes and query-rate metrics remain unavailable.

Configuration uses the shared database service boundary:

```toml
[[services]]
name = "database"
type = "db"
provider = "postgres"
address = "postgres://user:password@localhost:5432/database?sslmode=disable"
```
