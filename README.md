# Sen

Inspired by SenbonZakura, sen analyzes applications as collections of processes and supporting
services. It combines source analysis with live runtime and service telemetry
so performance costs and interactions can be visualized across files,
functions, call paths, and dependencies.

Sen currently supports Go, Node.js, and Rust applications plus Redis and PostgreSQL services,
combining:

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
./bin/sen run ./examples/rust
./bin/sen run ./examples/redis
./bin/sen run ./examples/postgres
./bin/sen run --config ./config/sen.toml
```

Use `-c` as the short form of `--config`.

Press `M` in a service workspace to open its runtime metrics dashboard.

## License

Sen is available under the [MIT License](LICENSE).
