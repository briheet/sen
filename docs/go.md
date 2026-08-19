# Using sen with Go

sen expects each Go service path to point to a buildable `main` package. A typical project looks like:

```text
my-backend/
├── sen.toml
├── go.mod
└── cmd/
    ├── api/main.go
    └── worker/main.go
```

Add each process to `sen.toml`:

```toml
[project]
name = "my-backend"

[[services]]
name = "api"
type = "server"
lang = "go"
path = "./cmd/api"
build_args = ["-tags=production"]
run_args = ["--port", "8080"]

[[services]]
name = "worker"
type = "server"
lang = "go"
path = "./cmd/worker"

[[services]]
name = "cache"
type = "kv"
provider = "redis"
address = "localhost:6379"
```

`type` selects the TUI model, while `lang` or `provider` selects its implementation. Paths are relative to `sen.toml`. `build_args` are passed to Go package analysis and `go build`; `run_args` are passed to the compiled service.

Run every configured service concurrently:

```sh
sen run
sen run ./path/to/project
sen run -c ./path/to/sen.toml
```

sen instruments temporary builds without changing the project source. Service output is written to the run's `engine.log`, keeping the terminal available for the TUI.
