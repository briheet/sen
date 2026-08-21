# Sen

Inspired by SenbonZakura, sen maps application code and backing services into a live TUI.
It combines source analysis with runtime telemetry so you can see where time,
memory, and activity are showing up across functions, files, calls, and dependencies.

Sen currently supports Go and Node.js applications plus Redis and PostgreSQL services:

- Static call-graph analysis
- Process and memory measurements
- CPU profiles and runtime traces
- Source-level mapping into a TUI-owned `RuntimeGraph`

Redis workspaces show per-command activity; PostgreSQL workspaces show
per-statement and per-table activity. Both include provider-specific live
metrics dashboards.

The target source is not modified. Adapters instrument each runtime and remove
their temporary artifacts on exit.

## Usage

Go 1.25.12 or newer is required to build sen.

```sh
go build -o bin/sen ./cmd/sen
./bin/sen run
./bin/sen run ./examples/go/http
./bin/sen run ./examples/go/postgresredis
./bin/sen run ./examples/redis
./bin/sen run ./examples/postgres
./bin/sen run --config ./config/sen.toml
```

Use `-c` as the short form of `--config`.

Set a built-in theme by slug in `sen.toml`; omitting it uses `zakura`:

```toml
[project]
name = "my-backend"
theme = "catppuccin-mocha"
```

Use `h` / `l` or the arrow keys to switch between service tabs.
Press `M` in a service workspace to open its runtime metrics dashboard.

## License

Sen is available under the [MIT License](LICENSE).
