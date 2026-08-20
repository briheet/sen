# Rust example

This profile demonstrates Sen's native Rust profiler and optional Tokio Console metrics. Sen builds into an isolated Cargo target directory and never edits this source tree.

```sh
go build -ldflags='-s -w' -o bin/sen ./cmd/sen/main.go
DEBUG=true ./bin/sen run --profile examples/rust
```

Generate load in another terminal:

```sh
while true; do curl -s http://127.0.0.1:8081/work >/dev/null; done
```

Press `M` in Sen to open the Rust metrics dashboard. Change `tokio_console` to `"off"` for native profiling alone, or to `"existing"` when the application already calls `console_subscriber::init()`.
