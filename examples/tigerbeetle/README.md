# TigerBeetle example

This example uses TigerBeetle's native binary. Sen only listens for metrics;
it does not start, stop, or modify the cluster, and the observed application
does not need a Sen library.

## 1. Install TigerBeetle

Follow the official install instructions and place `tigerbeetle` on your
`PATH`. For example, verify it with:

```sh
tigerbeetle version
```

## 2. Format and start a development replica

Formatting is required only once. The data file lives outside this repository:

```sh
mkdir -p /tmp/sen-tigerbeetle
tigerbeetle format \
  --cluster=0 \
  --replica=0 \
  --replica-count=1 \
  --development \
  /tmp/sen-tigerbeetle/0_0.tigerbeetle
```

Start the replica with StatsD telemetry directed at Sen:

```sh
tigerbeetle start \
  --addresses=127.0.0.1:3000 \
  --development \
  --experimental \
  --statsd=127.0.0.1:8125 \
  /tmp/sen-tigerbeetle/0_0.tigerbeetle
```

`metrics_address` must match `--statsd`. For a replicated cluster, keep
`addresses` in replica-index order, exactly like TigerBeetle's `--addresses`.

## 3. Run Sen

From the repository root:

```sh
DEBUG=true ./bin/sen run --profile examples/tigerbeetle
```

TigerBeetle emits metrics in roughly ten-second windows, so the first graph
activity is not immediate. Press `M` for cluster metrics and use the two dots
below the graph to switch between operations and replicas.

## 4. Generate activity

In another terminal, run the included REPL workload:

```sh
tigerbeetle repl --cluster=0 --addresses=127.0.0.1:3000 \
  < examples/tigerbeetle/workload.tigerbeetle
```

Repeated runs still exercise the request paths even when create calls report
that the fixed example IDs already exist.

TigerBeetle monitoring is currently experimental. See the official
[monitoring](https://docs.tigerbeetle.com/operating/monitoring/) and
[quick-start](https://docs.tigerbeetle.com/start/) documentation for current
CLI and deployment guidance.
