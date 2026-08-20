# Using sen with Rust

Sen supports executable Cargo targets on macOS and Linux, on both x86-64 and ARM64. It builds the selected target in an isolated temporary directory with frame pointers and DWARF symbols, then profiles the process with macOS `sample` or Linux `perf`. The application source and its normal Cargo target directory are not changed.

```toml
[project]
name = "rust-backend"

[[services]]
name = "api"
type = "server"
lang = "rust"
path = "."
build_args = ["--bin", "api"]
run_args = ["--port", "8080"]
tokio_console = "off"
```

`build_args` are passed to `cargo build`. A package or workspace with multiple executable targets must select one using `--bin`. Native process and CPU-profile telemetry requires no Rust dependency or source change.

## Optional Tokio metrics

Tokio task, poll, wake, busy-time, resource, and async-operation metrics use the official Tokio Console protocol. Add the official dependencies to the application:

```toml
[dependencies]
console-subscriber = "0.5"
tokio = { version = "1", features = ["tracing"] }
```

Then choose one explicit mode:

- `tokio_console = "inject"` makes Sen insert `console_subscriber::init()` into a temporary mirrored crate root. It does not edit the project.
- `tokio_console = "existing"` connects to an initialization already present in the application.
- `tokio_console = "off"` keeps native profiling only and is the default.

Configured Tokio modes fail startup if the subscriber cannot be reached, preventing a dashboard that silently appears healthy without data. Sen chooses a loopback port for each process and supplies `TOKIO_CONSOLE_BIND` to the child.

On Linux, `perf` must be installed and the current user must be allowed to profile the child process. On macOS, Sen invokes the system `sample` command. Missing tools or OS permissions fail startup with the profiler error.
