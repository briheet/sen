package console

import (
	"testing"
	"time"

	asyncopspb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/asyncops"
	commonpb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/common"
	instrumentpb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/instrument"
	resourcespb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/resources"
	taskspb "github.com/briheet/sen/internal/adapters/rust/runtime/console/proto/tasks"
	durationpb "github.com/golang/protobuf/ptypes/duration"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/require"
)

func TestSnapshotAggregatesLatestConsoleStats(t *testing.T) {
	t.Parallel()
	collector := &Collector{
		tasks: make(map[uint64]*taskspb.Stats), resources: make(map[uint64]*resourcespb.Stats), asyncOps: make(map[uint64]*asyncopspb.Stats),
	}
	collector.apply(&instrumentpb.Update{
		TaskUpdate: &taskspb.TaskUpdate{DroppedEvents: 2, StatsUpdate: map[uint64]*taskspb.Stats{
			1: {Wakes: 8, SelfWakes: 3, ScheduledTime: &durationpb.Duration{Nanos: 4}, PollStats: &commonpb.PollStats{Polls: 5, BusyTime: &durationpb.Duration{Nanos: 7}}},
			2: {DroppedAt: &timestamppb.Timestamp{Seconds: 1}},
		}},
		ResourceUpdate: &resourcespb.ResourceUpdate{DroppedEvents: 3, StatsUpdate: map[uint64]*resourcespb.Stats{1: {}, 2: {DroppedAt: &timestamppb.Timestamp{Seconds: 1}}}},
		AsyncOpUpdate:  &asyncopspb.AsyncOpUpdate{DroppedEvents: 4, StatsUpdate: map[uint64]*asyncopspb.Stats{1: {}}},
	})

	metrics := collector.Snapshot()
	require.True(t, metrics.TokioEnabled)
	require.Equal(t, uint64(1), metrics.LiveTasks)
	require.Equal(t, uint64(2), metrics.TotalTasks)
	require.Equal(t, uint64(1), metrics.CompletedTasks)
	require.Equal(t, uint64(5), metrics.Polls)
	require.Equal(t, uint64(8), metrics.Wakes)
	require.Equal(t, uint64(3), metrics.SelfWakes)
	require.Equal(t, 7*time.Nanosecond, metrics.BusyTime)
	require.Equal(t, 4*time.Nanosecond, metrics.ScheduledTime)
	require.Equal(t, uint64(1), metrics.LiveResources)
	require.Equal(t, uint64(1), metrics.LiveAsyncOperations)
	require.Equal(t, uint64(9), metrics.DroppedEvents)
}
