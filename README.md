# Senbon

Senbon analyzes an application's source code and combines it with live runtime
data so performance costs can be displayed on files, functions, and call paths.

Go is the first supported target. Senbon currently combines:

- SSA and reachable call-graph analysis
- `runtime/metrics` process measurements
- pprof CPU and allocation stacks
- Go runtime trace events
- Source-level mapping into a TUI-owned `RuntimeGraph`

The target source is not modified. Senbon builds it with an overlay, runs an
injected collector over a local Unix socket, and removes temporary artifacts on
exit.

## Usage

Go 1.25.12 or newer is required.

```sh
go build -o bin/senbon ./cmd/senbon
./bin/senbon run ./tests/examples
```

The CLI currently runs the target and prints a collection summary. The TUI is
under development.

## License

Senbon is available under the [MIT License](LICENSE).
