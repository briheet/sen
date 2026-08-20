# Redis example

Start the disposable Redis container, then run sen from the repository root:

```sh
docker compose -f ./examples/redis/compose.yaml up -d
./bin/sen run ./examples/redis
```

Select `cache` in the TUI. Press `M` to open the Redis metrics dashboard.

Stop and remove the container with:

```sh
docker compose -f ./examples/redis/compose.yaml down
```

Change `address` in `sen.toml` if Redis is listening elsewhere. The current
adapter expects a TCP address in `host:port` form and does not yet expose
authentication or TLS settings.
