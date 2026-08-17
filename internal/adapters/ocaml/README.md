# OCaml Adapter

Senbon's OCaml integration lives in `internal/adapters/ocaml/`. It follows the
same pattern as the Go/Node/Zig adapters: an `Analyzer` builds a normalized
static graph from the target source, and a `Runtime` collects runtime data and
converts it into `model.Observation`.

OCaml is a *functional* language, so this adapter demonstrates Senbon on an FP
codebase: module-style top-level value bindings, recursion, and higher-order
composition map naturally onto the call graph.

## Layout

```
internal/adapters/ocaml/
├── ocaml.go              Adapter (Analyze + Open), factory-registered as "ocaml"
├── analysis/
│   ├── analyze.ml        (//go:embed) compiler-libs AST walker → JSON
│   ├── analysis.go       ocamlc-compile helper → Project{Graph, Entry}
│   ├── analysis_test.go
│   └── analysis_benchmark_test.go
└── runtime/
    ├── runtime.go        Start/Collect/Wait/Stop/Cleanup → model.Observation
    ├── process.go        ocamlc build + launch with runtime-events env
    ├── events.go         (.events ring buffer) → GC collection counts
    ├── events_test.go / events_benchmark_test.go
    └── runtime_test.go   ocamlc-gated integration test
```

## How it works

### 1. Static analysis (`analysis/`)

`analyze.ml` is compiled on the fly against **OCaml's own compiler-libs** (the
real `Parser` + `Ast_iterator`), so the callgraph is accurate — not a
regex/tokenizer heuristic.

It walks the `Parsetree` and:
- Collects top-level `let`/`let rec` **value bindings** as function nodes.
- Records **call sites** (`Pexp_apply`) attributed to the enclosing function.
- Emits JSON `{entry, functions:[...], edges:[{from,to}]}`.

Operator calls (`+`, `-`, `<`, ...) are filtered out so only user-defined
function edges remain. Node ids are contiguous in first-binding order.

### 2. Per-call instrumentation (`runtime/process/instrument.ml`)

`instrument.ml` uses OCaml's `Ast_mapper` to rewrite each top-level function

    let f P1..Pn = body

into

    let f P1..Pn = Profiler.__p_incr <id> (fun () -> body)

so every invocation of `f` emits a `Runtime_events` custom span carrying the
function's id. The rewrite happens **at build time on a copy of the source** —
the user's `.ml` files are never modified. `profiler.ml` registers a
`Runtime_events` int-span and writes Begin/End per call; the runtime writes
these into the same `.events` ring on exit.

`instrument.ml` also emits `fns.json` mapping `id -> function name`, which
senbon reads to attribute spans to functions.

### 3. Runtime (`runtime/`)

The target is compiled with `ocamlopt` linking the embedded `profiler.ml` and
`runtime_events.cmxa`, then launched with OCaml's runtime-events tracing enabled
entirely via environment variables — no source modification:

```go
OCAML_RUNTIME_EVENTS_START=1
OCAML_RUNTIME_EVENTS_DIR=<dir>
OCAML_RUNTIME_EVENTS_PRESERVE=1
```

At exit the runtime writes a `<pid>.events` ring buffer.

### 4. Decoding (`runtime/events.go`)

`decodeEvents` parses the `.events` ring buffer (documented in
`caml/runtime_events.h`):

- Reads the `runtime_events_metadata_header` (version, ring offsets).
- Walks packed 64-bit items; `EV_EXIT` span endings yield `EV_MINOR`/`EV_MAJOR`
  GC collection counts; **user custom-span items** (bit 53) yield per-function
  `{timestamp, function_id}` observations.
- Returns `eventCounts{Minor, Major, FunctionSpans, Spans}`.

`runtime.Collect` maps these onto a `model.Profile` (name `"ocaml"`,
`calls:count` samples, one location per function) and a `model.Trace` timeline
of per-function stack-samples, giving the dot-map real per-function heat and
pulses.

The `.events` ring is only flushed to disk on process exit, so `Collect` waits
for the target to finish before reading (like the Zig adapter's capture-then-
replay model).

## Notes & limitations

- **Per-function tracing is self-contained** (no perf/Instruments/root),
  achieved via build-time AST instrumentation, like the Zig SIGPROF walker.
- **Native optimizations may inline small functions** (e.g. `ocamlopt` can
  inline a trivial helper), so their spans may not emit — this is inherent to
  native compilation, not a decoder gap.
- **Runtime overhead**: one custom-span write per function call, negligible for
  profiling but measurable on extremely hot micro-loops.
- GC/allocation counters come from `Runtime_events`; deep per-function stack
  *sampling* (as in Go/Zig) is not available — the timeline is span-based.

## Tests & benchmarks

- `go test ./internal/adapters/ocaml/...` — static-graph parse tests, an
  `ocamlopt`-gated integration `TestCollect` (builds a real instrumented target
  and collects GC + per-function data; skips without `ocamlopt`), and an
  env-gated `.events` decode test.
- `go test -bench=. -benchmem ./internal/adapters/ocaml/...` — `decodeEvents`
  (~1.3µs for 1000 items, ~1 alloc/op) and analysis `parse` (~1.2ms for 2.4k
  functions).

## Example

```
senbon run ocaml ./examples/ocaml
```

`examples/ocaml/calc.ml` is a small functional pipeline (`mix → fib/add`
composition) exercising recursion and top-level binding analysis.
