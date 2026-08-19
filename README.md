![Senbon Ascii](./assets/senbon_ascii.png)

Inspired by SenbonZakura, sen analyzes applications as collections of processes and supporting
services. It combines source analysis with live runtime and service telemetry
so performance costs and interactions can be visualized across files,
functions, call paths, and dependencies.

Sen currently supports Go and Node.js, combining:

- Static call-graph analysis
- Process and memory measurements
- CPU profiles and runtime traces
- Source-level mapping into a TUI-owned `RuntimeGraph`

Project configuration can also describe supporting services such as Redis,
with additional datastore integrations, including PostgreSQL, being added.

The target source is not modified. Adapters instrument each runtime and remove
their temporary artifacts on exit.

## Usage

Go 1.25.12 or newer is required to build sen.

```sh
go build -o bin/sen ./cmd/sen
./bin/sen run
./bin/sen run ./examples/go/http
./bin/sen run --config ./config/sen.toml
```

Use `-c` as the short form of `--config`.

The CLI currently runs and samples configured targets. The TUI is under development.

## License

Sen is available under the [MIT License](LICENSE).
