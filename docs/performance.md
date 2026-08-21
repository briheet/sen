# Performance work

We profiled sen under sustained Go, PostgreSQL, and Redis load, then focused on
the two largest sources of work: trace ingestion and TUI redraws.

## What changed

- **Go traces:** `internal/adapters/golang/runtime/trace` now streams trace
  events into compact aggregates instead of retaining every event for each
  collection window. Stack and event buffers are reused where possible.
- **Runtime model:** `internal/model` consumes those aggregates directly and
  avoids rebuilding repeated stack data.
- **TUI updates:** `internal/tui` now uses event-driven telemetry, precise
  revisions, cached page output, reusable overlay canvases, and pointer-backed
  page models. Hidden or unchanged metrics no longer trigger full redraws.

## Results

The hard-load profiles ran for different lengths, so rates are the fairest
comparison.

| Measurement | Before | After | Change |
| --- | ---: | ---: | ---: |
| Total allocation rate | 239 MB/s | 8.47 MB/s | **28x lower** |
| Trace allocation rate | 231 MB/s | 5.23 MB/s | **44x lower** |
| Allocation frequency | 47.9k objects/s | 33.2k objects/s | **31% lower** |
| CPU utilization | 20.49% | 14.10% | **31% lower** |
| Live heap after TUI caching | 35.1 MiB | 20.5 MiB | **40% lower** |
| Screen-buffer allocation | 315.8 MB | 5.1 MB | **62x lower** |

In cumulative `alloc_space`, the comparable hard-load run fell from about
104 GB to 3.1 GB. The trace decoder alone fell from 100.6 GB to 1.9 GB. No-op
TUI updates, which previously allocated 18,432 B/op, now allocate 0 B/op.

The main remaining costs are trace-reader buffer growth and graph image
rendering. These results came from `go tool pprof` CPU and heap profiles, with
benchmarks used for the no-op update and repeated-stack paths.
