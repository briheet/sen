# TigerBeetle adapter

The adapter attaches to TigerBeetle's experimental DogStatsD stream. It does
not start the cluster and does not link a TigerBeetle client library.

`analysis` creates stable operation and replica topology from the ordered
configured addresses. `runtime` binds the configured UDP endpoint, groups the
multi-datagram ten-second emission into one observation, and maps public
request operations back to the synthetic graph.

TigerBeetle must be started with:

```text
--experimental --statsd=<metrics_address>
```

One Sen service represents one cluster. Packets for any additional cluster on
the same listener are ignored.
