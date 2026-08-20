# PostgreSQL example

Start PostgreSQL with `pg_stat_statements` enabled:

```sh
docker compose -f examples/postgres/compose.yaml up -d
```

Run sen from the repository root:

```sh
DEBUG=true ./bin/sen run --profile examples/postgres
```

In another terminal, generate query and table activity:

```sh
docker compose -f examples/postgres/compose.yaml exec postgres \
  pgbench -U sen -d sen -c 8 -j 4 -T 60 -n -f /workload.sql
```

The first graph page shows SQL statements and the second shows tables; click
the dots below the graph to switch. Press `M` for PostgreSQL metrics. Stop the
database with:

```sh
docker compose -f examples/postgres/compose.yaml down
```
