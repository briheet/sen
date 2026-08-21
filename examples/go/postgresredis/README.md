# Go + PostgreSQL + Redis example

Start PostgreSQL and Redis:

```sh
docker compose -f examples/go/postgresredis/compose.yaml up -d
```

Run sen from the repository root:

```sh
DEBUG=true ./bin/sen run --profile examples/go/postgresredis
```

In another terminal, drive mixed application, database, and cache activity:

```sh
python3 examples/go/postgresredis/load.py
```

The default workload runs for five minutes with 128 workers and no client-side
rate cap. Use `--rate` to cap requests per second or `--workers` to change
concurrency.

Select `api` for the Go server graph, `database` for PostgreSQL statements and
tables, and `cache` for Redis commands. Press `M` in the PostgreSQL or Redis
workspaces to open provider metrics.

Stop the disposable services with:

```sh
docker compose -f examples/go/postgresredis/compose.yaml down
```
