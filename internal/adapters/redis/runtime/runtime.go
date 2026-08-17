// Package runtime collects metrics and per-command heat from a running Redis.
package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/redis/runtime/metrics"
	"github.com/briheet/sen/internal/adapters/redis/runtime/trace"
	"github.com/briheet/sen/internal/model"
	"github.com/redis/go-redis/v9"
)

// Collector owns a connection to a running Redis server.
type Collector struct {
	client *redis.Client

	Metrics *model.RuntimeMetrics
	Profile *model.Profile

	doneOnce sync.Once
	done     chan struct{}
}

var _ adapters.Runtime = (*Collector)(nil)

// NewCollector dials the given Redis address (e.g. "localhost:6379").
func NewCollector(addr string) *Collector {
	return &Collector{
		client: redis.NewClient(&redis.Options{
			Addr:        addr,
			DialTimeout: 2 * time.Second,
			MaxRetries:  2,
		}),
		done: make(chan struct{}),
	}
}

// Start verifies connectivity and enables per-command latency tracking using
// the running server (no source build or target launch is required).
func (c *Collector) Start(ctx context.Context) error {
	if _, err := c.client.Ping(ctx).Result(); err != nil {
		return err
	}
	_ = c.client.ConfigSet(ctx, "latency-tracking", "yes").Err()
	return nil
}

// Collect pulls one complete snapshot: server-wide metrics plus a per-command
// heat profile derived from INFO commandstats/latencystats. After the first
// successful snapshot the collector considers its single observation complete.
func (c *Collector) Collect(ctx context.Context) (model.Observation, error) {
	body, err := c.client.Info(ctx, "memory", "stats", "cpu", "clients").Result()
	if err != nil {
		return model.Observation{}, err
	}
	cmdstats, err := c.client.Info(ctx, "commandstats").Result()
	if err != nil {
		return model.Observation{}, err
	}

	c.Metrics = metrics.Decode(body)
	c.Profile = trace.Decode(cmdstats)

	profiles := map[string]*model.Profile{}
	if len(c.Profile.Samples) > 0 {
		profiles[trace.ProfileName] = c.Profile
	}
	c.finish()
	return model.Observation{Metrics: c.Metrics, Profiles: profiles}, nil
}

// Wait blocks until the first snapshot is collected or Stop is called.
func (c *Collector) Wait() error {
	<-c.done
	return nil
}

// Stop terminates observation and releases the connection.
func (c *Collector) Stop() error {
	c.finish()
	return c.client.Close()
}

// Cleanup releases the connection.
func (c *Collector) Cleanup() error {
	c.finish()
	return c.client.Close()
}

// finish unblocks any waiter exactly once.
func (c *Collector) finish() {
	c.doneOnce.Do(func() { close(c.done) })
}
