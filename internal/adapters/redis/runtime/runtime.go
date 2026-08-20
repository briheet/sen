// Package runtime observes a running Redis server over its public protocol.
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

const (
	collectionInterval = time.Second
	dialTimeout        = 2 * time.Second
)

// Collector owns the Redis connection and the state needed to turn cumulative
// commandstats counters into per-window profiles.
type Collector struct {
	client *redis.Client

	lastCollection time.Time
	commandStats   trace.Snapshot

	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

var _ adapters.Runtime = (*Collector)(nil)

// NewCollector prepares a client for addr, for example "localhost:6379".
// Connectivity is verified by Start so engine construction remains side-effect
// free and produces a useful runtime error at the normal lifecycle boundary.
func NewCollector(addr string) *Collector {
	return &Collector{
		client: redis.NewClient(&redis.Options{
			Addr:        addr,
			DialTimeout: dialTimeout,
			MaxRetries:  2,
		}),
		done: make(chan struct{}),
	}
}

// Start verifies that the configured Redis server is reachable. Sen only
// reads INFO and PING; it does not change server configuration.
func (c *Collector) Start(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Collect reads one INFO snapshot and returns both server metrics and command
// activity. Collection is paced here because Redis replies immediately, unlike
// the one-second profile windows used by the Go and Node adapters.
func (c *Collector) Collect(ctx context.Context) (model.Observation, error) {
	if err := c.waitForWindow(ctx); err != nil {
		return model.Observation{}, err
	}
	// The default INFO response omits commandstats on current Redis releases.
	// Request all sections so the same snapshot feeds both metrics and activity.
	body, err := c.client.Info(ctx, "all").Result()
	if err != nil {
		return model.Observation{}, err
	}

	collectedAt := time.Now()
	duration := collectionInterval
	if !c.lastCollection.IsZero() {
		duration = collectedAt.Sub(c.lastCollection)
	}
	currentStats := trace.Parse(body)
	var window trace.Snapshot
	if c.commandStats != nil {
		window = currentStats.Delta(c.commandStats)
	}
	c.commandStats = currentStats
	c.lastCollection = collectedAt

	var profiles map[string]*model.Profile
	if len(window) > 0 {
		profiles = map[string]*model.Profile{trace.ProfileName: window.Profile(duration)}
	}
	return model.Observation{
		Metrics:  metrics.Decode(body),
		Profiles: profiles,
	}, nil
}

func (c *Collector) waitForWindow(ctx context.Context) error {
	if c.lastCollection.IsZero() {
		return nil
	}
	delay := time.Until(c.lastCollection.Add(collectionInterval))
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Wait blocks for the lifetime of the attached service. Redis is external to
// sen, so only Stop or Cleanup completes this runtime.
func (c *Collector) Wait() error {
	<-c.done
	return nil
}

// Stop releases the Redis connection and unblocks Wait.
func (c *Collector) Stop() error { return c.close() }

// Cleanup is idempotent and releases any connection not already stopped.
func (c *Collector) Cleanup() error { return c.close() }

func (c *Collector) close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.client.Close()
		close(c.done)
	})
	return c.closeErr
}
