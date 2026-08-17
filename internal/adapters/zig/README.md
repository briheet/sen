# Zig Adapter

Senbon's Zig integration lives in `internal/adapters/zig/`. It follows the same
pattern as the Go and Node adapters: an `Analyzer` builds a normalized static
graph, and a `Runtime` collects live runtime samples and converts them into
`model.Profile` + `model.Trace` snapshots.

The headline property: **zero modification of the target's source code.** Senbon
compiles the program with a sampler *wrapper* as the main module, exactly like
the Go overlay trick.

## Layout

```
internal/adapters/zig/
├── zig.go                  Adapter (Analyze + Open), factory-registered as "zig"
├── analysis/
│   ├── analyze.zig         (//go:embed) tokenizer-based analyzer → JSON
│   ├── analysis.go         zig run → Project{Graph, Entry, Modules, Imports}
│   ├── analysis_test.go
│   └── analysis_benchmark_test.go
└── runtime/
    ├── runtime.go          Start/Collect/Wait/Stop/Cleanup → model.Observation
    ├── process/
    │   ├── process.go      zig build-exe module graph; dsymutil; stop/kill
    │   ├── sampler.zig     (//go:embed) wrapper main: SIGPROF → samples file
    │   └── process_test.go
    ├── decode.go           samples → model.Profile + model.Trace
    ├── symbolize.go        DWARF (elf/macho + dSYM) PC→file:line
    ├── decode_test.go
    ├── decode_benchmark_test.go
    └── runtime_test.go     zig-gated integration test
```

## How it works

### 1. Static analysis (`analysis/`)

`analyze.zig` uses Zig's own tokenizer (`std.zig.Tokenizer`) — it already skips
comments and strings, making lexical analysis reliable. It walks each `.zig`
file and emits JSON:

- **Functions**: `fn` declarations (name token after `fn`) with start/end spans.
- **Calls**: `(` tokens preceded by an identifier, attributed to the enclosing
  function by brace depth.
- **Imports**: `@import("...")` builtin calls, producing per-file import lists.
- **Edges**: inner calls resolved to declared function ids (same file first,
  then project-wide).

Run as a subprocess: `zig run analyze.zig -lc -- <projectDir> <outputPath>`.

`analysis.go` maps the JSON into `model.StaticGraph` and also keeps the module
map (`import name → source file`) and per-file import lists needed to build the
instrumented target.

### 2. Runtime injection (`runtime/process/`)

The target is rebuilt with the sampler as the root module:

```
zig build-exe -lc \
  --dep user=user \
  -Mroot=<sampler.zig> \
  <--dep mod=mod for each import> \
  -Muser=<entry.zig> \
  -M<module>=<file.zig> ... \
  -femit-bin=<out> -O Debug
```

`s/--dep` semantics: `--dep` must precede the `-M` it applies to, and its value
is a **module name** (defined by a later `-M`), not a path. `-Mroot` defines the
main module; Zig special-cases `@import("root")` to it regardless of name.

The target is built with `-O Debug` (frame pointers + DWARF) and linked with
libc (`-lc`, required for `setitimer`/fd I/O).

On macOS, DWARF is extracted into a dSYM bundle via `dsymutil` (Zig's Mach-O
executables use a debug map rather than `__DWARF` section).

### 3. Sampling (`runtime/process/sampler.zig`)

The wrapper main (built as `root`) runs the sampler, then invokes
`@import("user").main()`:

- `setitimer(ITIMER_PROF, 1ms)` delivers SIGPROF on a fixed interval.
- The handler unwinds the interrupted stack via the kernel ucontext
  (`std.debug.cpu_context.fromPosixSignalContext`) with a fast **frame-pointer
  walk** — no DWARF, no malloc, async-signal-safe.
- Each sample is `write(2)`-ed directly as a line: `<leaf pc> <frame pc> ...`.
- A `senbon_marker` export records the linked address so Go can compute the
  ASLR slide at runtime.

Two bugs were found and fixed during development:

- **Signal-stealing collector thread.** macOS delivers a process-directed
  SIGPROF to an arbitrary thread including a sleeping collector thread, so all
  samples landed in its `nanosleep`. Fix: drop the collector thread entirely;
  the handler writes synchronously.
- **FP-walk runaway.** Reading garbage frame pointers produced unbounded stacks.
  Fix: the next frame pointer must be strictly greater than the current one,
  else stop.

### 4. Decoding + symbolization (`runtime/`)

`Collect` reads the samples file after the window and:

1. **`symbolizer`** — computes the ASLR slide by subtracting the linked
   `senbon_marker` address (from the binary symtab) from the runtime base in the
   header, then resolves PCs to `file:line` via DWARF line tables
   (`debug/elf` + `debug/macho`, reading the dSYM when present). Unresolvable
   PCs fall through to unmapped (no location).
2. **`decodeSamples`** — converts lines into `model.Profile` (one sample per
   tick, leaf-first stack, `cpu:nanoseconds` values) and `model.Trace`
   (`EventStackSample` per sample at cumulative time, deduplicated stacks).
   Interval comes from the file header.

## Key design decisions

- **ASLR slide via marker**: `-fno-PIE` does not disable ASLR on macOS, so the
  marker/pc-slide approach is required instead of treating raw PCs as static.
- **NodeID offset +1**: Zig analyzer IDs start at 0, but `model`'s frame mapper
  treats `0` as "not found". `parse()` offsets all IDs by +1 so the first
  function (`fib`, typically) isn't silently dropped.
- **[main] signature (v1)**: supports `pub fn main()` with `void`/`!void`/`u8`
  returns and no parameters; other signatures are a clear compile error.
- **Sampling only**: no process metrics in v1 (empty `model.RuntimeMetrics`).

## Limitations

- Stacks are effectively depth-1 in optimized builds: the FP walk resolves the
  hot **leaf** function reliably (enough for heat and pulse visuals) but caller
  frames above it can hit non-unwindable Zig/std frames.
- Requires `zig` and `dsymutil` (macOS) on the PATH.
- Analyzer std APIs (`std.zig.Tokenizer`) carry the std's "no API guarantees"
  caveat — pinned to Zig 0.16.0.

## Tests & benchmarks

- `go test ./internal/adapters/zig/...` — unit tests for analysis parse, decode,
  header parse, and the module-graph build args; the integration `TestCollect`
  builds a real target, samples it, and asserts DWARF maps samples to the target
  source (skipped when `zig` is absent).
- `go test -bench=. -benchmem ./internal/adapters/zig/...` — `decodeSamples`
  (~368µs for 1000 samples) and analysis `parse` (~2.5ms for 2.4k functions),
  both one-shot/leaf-bound.

## Example

```
senbon run zig ./examples/zig
```

`examples/zig/` is a compute pipeline (`main → router.run → services.mix/fib`)
in three modules, exercising multi-module injection.
