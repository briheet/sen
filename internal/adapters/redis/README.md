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

## How it works

- **Metrics** come from `INFO memory stats cpu clients`, folded into
  `model.RuntimeMetrics` (heap/RSS/live dataset, user/sys CPU, connected
  clients, processed commands, accepted connections).
- **Tracing / heat** comes from `INFO commandstats` (`cmdstat_<cmd>:
  calls=...,usec=...`) and is normalized into a per-command profile whose
  samples attribute `calls` and `usec` to the matching synthetic command nodes.
  This is the low-overhead analogue of SLOWLOG-based slow-command tracing and
  needs no injection into the server.

`Start` verifies connectivity with `PING` and enables `latency-tracking` so
`latencystats` are populated. `Collect` returns one snapshot: global metrics
plus a per-command heat profile.

## Synthetic graph

Redis has no source tree to analyze, so `analysis` synthesizes a `StaticGraph`
with one node per well-known command (plus a `redis-server` root). Observed
per-command heat is attributed onto these nodes, giving the TUI a stable
surface across snapshots. Unknown commands are filtered out of attribution but
remain countable through the unmapped pool.

Files:

- `redis.go` — `Application` adapter (analyze + open).
- `analysis/` — synthetic command graph.
- `runtime/` — live collector; `metrics/` parses INFO, `trace/` builds the
  per-command profile.
