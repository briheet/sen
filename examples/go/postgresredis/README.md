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
python3 examples/go/postgresredis/requests.py --duration 45 --rate 20
```

Select `api` for the Go server graph, `database` for PostgreSQL statements and
tables, and `cache` for Redis commands. Press `M` in the PostgreSQL or Redis
workspaces to open provider metrics.

Stop the disposable services with:

```sh
docker compose -f examples/go/postgresredis/compose.yaml down
```
