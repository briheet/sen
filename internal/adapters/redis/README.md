# Redis adapter

Observes a **running** Redis server over its public protocol. Unlike the Go and
Node adapters, there is nothing to build or launch: sen attaches to an
already-running instance by address and pulls telemetry from it.

```toml
[[services]]
name = "cache"
type = "kv"
provider = "redis"
address = "127.0.0.1:6379"
```

## Collection pipeline

- `Start` uses `PING` to verify the configured address. The adapter is
  read-only and does not change Redis configuration.
- `Collect` reads one complete `INFO` response every second.
- `runtime/metrics` decodes server, memory, CPU, client, network, keyspace, and
  command counters into `model.RedisMetrics`.
- `runtime/trace` converts cumulative `cmdstat_*` values into per-window
  profiles. Calls and command time are attributed to synthetic command nodes.
- `Wait` remains active until sen stops the attachment, allowing Redis-only
  projects to keep a live TUI open.

No source injection, `MONITOR`, `SLOWLOG`, or server-side configuration changes
are required.

## Synthetic graph

Redis has no source tree to analyze, so `analysis` synthesizes a `StaticGraph`
with one node per well-known command (plus a `redis-server` root). Observed
per-command heat is attributed onto these nodes, giving the TUI a stable
surface across snapshots. Unknown commands are ignored because they have no
stable node in the synthetic graph.

Files:

- `redis.go` — small adapter boundary used by the engine factory.
- `analysis/` — deterministic synthetic command graph.
- `runtime/` — connection lifecycle and one-second collection loop.
- `runtime/metrics/` — INFO metrics decoding.
- `runtime/trace/` — commandstats parsing, deltas, and normalized profiles.
