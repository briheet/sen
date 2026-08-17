# Adapters

Adapters translate target-specific analysis and telemetry into `model` types.

- `Application` analyzes source and opens its runtime.
- `Runtime` starts, samples, stops, and cleans up the target.

Each implementation lives in its own directory and declares conformance:

```go
var _ adapters.Application = (*Adapter)(nil)
var _ adapters.Runtime = (*Runtime)(nil)
```

`factory` selects the concrete adapter used by the engine.

Adapters may import `model`; `model` must not import an adapter.
