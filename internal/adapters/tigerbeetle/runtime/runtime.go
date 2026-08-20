// Package runtime collects TigerBeetle telemetry without owning its process.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/adapters/tigerbeetle/analysis"
	wire "github.com/briheet/sen/internal/adapters/tigerbeetle/runtime/metrics"
	"github.com/briheet/sen/internal/model"
)

const (
	quietWindow     = 100 * time.Millisecond
	pollInterval    = 250 * time.Millisecond
	emissionWindow  = 10 * time.Second
	maxDatagramSize = 64 << 10
	profileName     = "tigerbeetle.requests"
)

type requestKey struct {
	replica   uint32
	operation string
}

type requestWindow struct {
	count uint64
	minUS float64
	maxUS float64
	avgUS float64
	sumUS float64
}

// Collector binds TigerBeetle's StatsD UDP target and retains the latest
// gauges while emitting request activity once per native telemetry burst.
type Collector struct {
	address string
	output  io.Writer
	conn    net.PacketConn

	cluster    string
	release    uint64
	replicas   map[uint32]model.TigerBeetleReplicaMetrics
	lastWindow time.Time

	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

var _ adapters.Runtime = (*Collector)(nil)

// NewCollector creates a collector for a literal UDP listen address.
func NewCollector(address string, output io.Writer) *Collector {
	return &Collector{
		address: address, output: output, replicas: make(map[uint32]model.TigerBeetleReplicaMetrics),
		done: make(chan struct{}),
	}
}

// Start binds the metrics endpoint before TigerBeetle sends its next burst.
func (c *Collector) Start(_ context.Context) error {
	conn, err := net.ListenPacket("udp", c.address)
	if err != nil {
		return fmt.Errorf("tigerbeetle: listen statsd: %w", err)
	}
	c.conn = conn
	return nil
}

// Collect reads one complete TigerBeetle emission burst. Metrics are sent as
// several UDP datagrams, so a short quiet window defines the burst boundary.
func (c *Collector) Collect(ctx context.Context) (model.Observation, error) {
	if c.conn == nil {
		return model.Observation{}, errors.New("tigerbeetle: collector not started")
	}
	buffer := make([]byte, maxDatagramSize)
	requests := make(map[requestKey]*requestWindow)
	received, malformed, foreign := false, 0, 0
	for {
		deadline := time.Now().Add(pollInterval)
		if received {
			deadline = time.Now().Add(quietWindow)
		}
		if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
			deadline = contextDeadline
		}
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return model.Observation{}, err
		}
		count, _, err := c.conn.ReadFrom(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return model.Observation{}, ctx.Err()
			}
			var networkError net.Error
			if errors.As(err, &networkError) && networkError.Timeout() {
				if received {
					break
				}
				continue
			}
			return model.Observation{}, err
		}
		received = true
		for _, line := range strings.Split(string(buffer[:count]), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			sample, parseErr := wire.ParseLine(line)
			if parseErr != nil {
				malformed++
				continue
			}
			if c.cluster != "" && sample.Cluster != c.cluster {
				foreign++
				continue
			}
			if c.cluster == "" {
				c.cluster = sample.Cluster
			}
			c.apply(sample, time.Now(), requests)
		}
	}
	if malformed+foreign > 0 && c.output != nil {
		_, _ = fmt.Fprintf(c.output, "tigerbeetle: ignored metrics malformed=%d foreign_cluster=%d\n", malformed, foreign)
	}
	return c.observation(time.Now(), requests), nil
}

func (c *Collector) apply(sample wire.Sample, observedAt time.Time, requests map[requestKey]*requestWindow) {
	replica := c.replicas[sample.Replica]
	replica.ObservedAt = observedAt
	value := uint64(sample.Value)
	switch sample.Name {
	case "release":
		c.release = value
	case "replica_status":
		replica.Status = value
	case "replica_sync_stage":
		replica.SyncStage = value
	case "replica_view":
		replica.View = value
	case "replica_op":
		replica.Operation = value
	case "replica_op_checkpoint":
		replica.Checkpoint = value
	case "replica_commit_min":
		replica.CommitMin = value
	case "replica_commit_max":
		replica.CommitMax = value
	case "replica_pipeline_queue_length":
		replica.PipelineQueueLength = value
	case "journal_dirty", "replica_journal_dirty":
		replica.JournalDirty = value
	case "journal_faulty", "replica_journal_faulty":
		replica.JournalFaulty = value
	case "grid_blocks_acquired":
		replica.GridBlocksAcquired = value
	case "grid_blocks_missing":
		replica.GridBlocksMissing = value
	case "grid_cache_hits":
		replica.GridCacheHits = value
	case "grid_cache_misses":
		replica.GridCacheMisses = value
	case "table_count_visible":
		switch strings.ToLower(sample.Tree) {
		case "accounts", "account":
			replica.Accounts = value
		case "transfers", "transfer":
			replica.Transfers = value
		}
	default:
		c.applyRequest(sample, requests)
	}
	c.replicas[sample.Replica] = replica
}

func (c *Collector) applyRequest(sample wire.Sample, requests map[requestKey]*requestWindow) {
	operation, ok := analysis.NormalizeOperation(sample.Operation)
	if !ok {
		return
	}
	key := requestKey{replica: sample.Replica, operation: operation}
	window := requests[key]
	if window == nil {
		window = new(requestWindow)
		requests[key] = window
	}
	switch sample.Name {
	case "replica_request_us.count", "replica_request_us_count":
		window.count += uint64(sample.Value)
	case "replica_request_us.sum", "replica_request_us_sum":
		window.sumUS += sample.Value
	case "replica_request_us.avg", "replica_request_us_avg":
		window.avgUS = sample.Value
	case "replica_request_us.min", "replica_request_us_min":
		if window.minUS == 0 || sample.Value < window.minUS {
			window.minUS = sample.Value
		}
	case "replica_request_us.max", "replica_request_us_max":
		window.maxUS = max(window.maxUS, sample.Value)
	}
}

func (c *Collector) observation(collectedAt time.Time, requests map[requestKey]*requestWindow) model.Observation {
	window := emissionWindow
	if !c.lastWindow.IsZero() {
		window = collectedAt.Sub(c.lastWindow)
	}
	c.lastWindow = collectedAt
	metrics := &model.RuntimeMetrics{TigerBeetle: model.TigerBeetleMetrics{
		Cluster: c.cluster, Release: c.release, Window: window,
		Replicas:   make(map[uint32]model.TigerBeetleReplicaMetrics, len(c.replicas)),
		Operations: make(map[string]model.TigerBeetleOperationMetrics),
	}}
	for id, replica := range c.replicas {
		metrics.TigerBeetle.Replicas[id] = replica
	}
	for key, request := range requests {
		operation := metrics.TigerBeetle.Operations[key.operation]
		operation.Requests += request.count
		operation.LatencySum += microseconds(request.sumUS)
		minimum := microseconds(request.minUS)
		if operation.LatencyMin == 0 || minimum > 0 && minimum < operation.LatencyMin {
			operation.LatencyMin = minimum
		}
		operation.LatencyMax = max(operation.LatencyMax, microseconds(request.maxUS))
		metrics.TigerBeetle.Operations[key.operation] = operation
	}
	for operation, values := range metrics.TigerBeetle.Operations {
		if values.Requests > 0 && values.LatencySum > 0 {
			values.LatencyAvg = values.LatencySum / time.Duration(values.Requests)
		}
		metrics.TigerBeetle.Operations[operation] = values
	}
	profiles := requestProfile(collectedAt.Add(-window), window, requests)
	return model.Observation{Metrics: metrics, Profiles: profiles}
}

func microseconds(value float64) time.Duration {
	if value <= 0 {
		return 0
	}
	return time.Duration(value * float64(time.Microsecond))
}

// Wait blocks until Sen detaches; the external cluster has no owned lifetime.
func (c *Collector) Wait() error { <-c.done; return nil }

// Stop detaches from the UDP socket.
func (c *Collector) Stop() error { return c.close() }

// Cleanup is idempotent.
func (c *Collector) Cleanup() error { return c.close() }

func (c *Collector) close() error {
	c.closeOnce.Do(func() {
		if c.conn != nil {
			c.closeErr = c.conn.Close()
		}
		close(c.done)
	})
	return c.closeErr
}
