// Package console collects aggregate Tokio runtime telemetry from console-subscriber.
package console

import (
	"context"
	"fmt"
	"sync"
	"time"

	asyncopspb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/asyncops"
	instrumentpb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/instrument"
	resourcespb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/resources"
	taskspb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/tasks"
	"github.com/briheet/sen/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Collector owns one version-matched Tokio Console update stream.
type Collector struct {
	conn   *grpc.ClientConn
	ctx    context.Context
	cancel context.CancelFunc

	mu        sync.RWMutex
	tasks     map[uint64]*taskspb.Stats
	resources map[uint64]*resourcespb.Stats
	asyncOps  map[uint64]*asyncopspb.Stats
	archived  model.RustMetrics
	dropped   uint64
	err       error
}

// Connect waits for console-subscriber and validates the stream by receiving its first update.
func Connect(ctx context.Context, address string) (*Collector, error) {
	dialCtx, cancelDial := context.WithTimeout(ctx, 10*time.Second)
	defer cancelDial()
	conn, err := grpc.DialContext(dialCtx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to Tokio Console at %s: %w", address, err)
	}
	streamCtx, cancel := context.WithCancel(context.Background())
	stream, err := instrumentpb.NewInstrumentClient(conn).WatchUpdates(streamCtx, &instrumentpb.InstrumentRequest{})
	if err != nil {
		cancel()
		_ = conn.Close()
		return nil, fmt.Errorf("start Tokio Console update stream: %w", err)
	}

	first := make(chan struct {
		update *instrumentpb.Update
		err    error
	}, 1)
	go func() {
		update, recvErr := stream.Recv()
		first <- struct {
			update *instrumentpb.Update
			err    error
		}{update, recvErr}
	}()
	var update *instrumentpb.Update
	select {
	case result := <-first:
		if result.err != nil {
			cancel()
			_ = conn.Close()
			return nil, fmt.Errorf("receive Tokio Console update: %w", result.err)
		}
		update = result.update
	case <-dialCtx.Done():
		cancel()
		_ = conn.Close()
		return nil, fmt.Errorf("wait for first Tokio Console update: %w", dialCtx.Err())
	}

	collector := &Collector{
		conn: conn, ctx: streamCtx, cancel: cancel,
		tasks:     make(map[uint64]*taskspb.Stats),
		resources: make(map[uint64]*resourcespb.Stats),
		asyncOps:  make(map[uint64]*asyncopspb.Stats),
	}
	collector.apply(update)
	go collector.receive(stream)
	return collector, nil
}

func (c *Collector) receive(stream instrumentpb.Instrument_WatchUpdatesClient) {
	for {
		update, err := stream.Recv()
		if err != nil {
			if c.ctx.Err() == nil {
				c.mu.Lock()
				c.err = fmt.Errorf("Tokio Console update stream: %w", err)
				c.mu.Unlock()
			}
			return
		}
		c.apply(update)
	}
}

// Err reports an unexpected update-stream failure.
func (c *Collector) Err() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.err
}

func (c *Collector) apply(update *instrumentpb.Update) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if tasks := update.GetTaskUpdate(); tasks != nil {
		for _, task := range tasks.GetNewTasks() {
			id := task.GetId().GetId()
			if _, exists := c.tasks[id]; !exists {
				c.tasks[id] = nil
				c.archived.TotalTasks++
			}
		}
		for id, stats := range tasks.GetStatsUpdate() {
			_, exists := c.tasks[id]
			if !exists {
				c.archived.TotalTasks++
			}
			if stats.GetDroppedAt() == nil {
				c.tasks[id] = stats
				continue
			}
			archiveTask(&c.archived, stats)
			c.archived.CompletedTasks++
			delete(c.tasks, id)
		}
		c.dropped += tasks.GetDroppedEvents()
	}
	if resources := update.GetResourceUpdate(); resources != nil {
		for _, resource := range resources.GetNewResources() {
			c.resources[resource.GetId().GetId()] = nil
		}
		for id, stats := range resources.GetStatsUpdate() {
			if stats.GetDroppedAt() == nil {
				c.resources[id] = stats
			} else {
				delete(c.resources, id)
			}
		}
		c.dropped += resources.GetDroppedEvents()
	}
	if asyncOps := update.GetAsyncOpUpdate(); asyncOps != nil {
		for _, asyncOp := range asyncOps.GetNewAsyncOps() {
			c.asyncOps[asyncOp.GetId().GetId()] = nil
		}
		for id, stats := range asyncOps.GetStatsUpdate() {
			if stats.GetDroppedAt() == nil {
				c.asyncOps[id] = stats
			} else {
				delete(c.asyncOps, id)
			}
		}
		c.dropped += asyncOps.GetDroppedEvents()
	}
}

// Snapshot returns cumulative Tokio measurements without exposing protocol types.
func (c *Collector) Snapshot() model.RustMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	metrics := c.archived
	metrics.TokioEnabled = true
	metrics.LiveTasks = uint64(len(c.tasks))
	metrics.DroppedEvents = c.dropped
	for _, stats := range c.tasks {
		archiveTask(&metrics, stats)
	}
	for _, stats := range c.resources {
		if stats.GetDroppedAt() == nil {
			metrics.LiveResources++
		}
	}
	for _, stats := range c.asyncOps {
		if stats.GetDroppedAt() == nil {
			metrics.LiveAsyncOperations++
		}
	}
	return metrics
}

func archiveTask(metrics *model.RustMetrics, stats *taskspb.Stats) {
	metrics.Wakes += stats.GetWakes()
	metrics.SelfWakes += stats.GetSelfWakes()
	metrics.ScheduledTime += duration(stats.GetScheduledTime())
	if polls := stats.GetPollStats(); polls != nil {
		metrics.Polls += polls.GetPolls()
		metrics.BusyTime += duration(polls.GetBusyTime())
	}
}

func duration(value interface {
	GetSeconds() int64
	GetNanos() int32
}) time.Duration {
	if value == nil {
		return 0
	}
	return time.Duration(value.GetSeconds())*time.Second + time.Duration(value.GetNanos())
}

// Close stops the update stream and releases its connection.
func (c *Collector) Close() error {
	c.cancel()
	return c.conn.Close()
}
