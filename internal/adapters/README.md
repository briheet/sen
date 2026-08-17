# Adapters

Adapters connect external targets to Senbon's language-neutral `model`.

Each adapter implements only the capabilities it supports from `adapter.go`:

- `Analyzer` builds a static source graph.
- `Runner` manages a target process.
- `MetricsCollector`, `Profiler`, and `Tracer` collect runtime data.

Implementations live in their own directory, such as `golang`, `python`, or
`redis`, and should include compile-time interface checks:

```go
var _ adapters.Tracer = (*Runtime)(nil)
```

Adapters may import `model`; `model` must not import an adapter.
