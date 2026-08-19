# Using sen with Node.js

sen expects each Node.js service path to contain a runnable entry file. It resolves `package.json` `main` or `bin`, then falls back to `index.js`, `index.mjs`, `index.cjs`, or `index.ts`.

```text
my-backend/
├── sen.toml
└── services/
    └── api/
        ├── package.json
        ├── index.js
        └── src/
```

Node.js and TypeScript `^5.9` are required. Install TypeScript in each analyzed project:

```sh
npm install --save-dev typescript@^5.9
```

Configure the service in `sen.toml`:

```toml
[project]
name = "my-backend"

[[services]]
name = "api"
type = "server"
lang = "node"
path = "./services/api"
build_args = ["--no-warnings"]
run_args = ["--port", "8080"]

[[services]]
name = "cache"
type = "kv"
provider = "redis"
address = "localhost:6379"
```

`type` selects the TUI model, while `lang` or `provider` selects its implementation. Paths are relative to `sen.toml`. Node.js has no sen build step: `build_args` are Node/V8 flags placed before the entry file, while `run_args` are application arguments placed after it.

Run every configured service concurrently:

```sh
sen run
sen run ./path/to/project
sen run -c ./path/to/sen.toml
```

sen starts Node.js with runtime instrumentation without changing the project source. Service output is written to the run's `engine.log`, keeping the terminal available for the TUI.
