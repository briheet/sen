package runtime

import (
	"testing"
	"time"

	"github.com/briheet/sen/internal/adapters/redis/analysis"
	"github.com/briheet/sen/internal/adapters/redis/runtime/trace"
	"github.com/briheet/sen/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservedAttribution(t *testing.T) {
	t.Parallel()

	cmdstats := "cmdstat_get:calls=3,usec=9,usec_per_call=3.00,rejected_calls=0,failed_calls=0\r\n" +
		"cmdstat_set:calls=5,usec=50,usec_per_call=10.00,rejected_calls=0,failed_calls=0\r\n"

	graph := analysis.BuildGraph()
	runtimeGraph := model.BuildRuntimeGraph(analysis.ModulePath, graph)
	profiles := map[string]*model.Profile{trace.ProfileName: trace.Parse(cmdstats).Profile(time.Second)}
	runtimeGraph.ApplyUpdate(runtimeGraph.BuildUpdate(&model.RuntimeMetrics{}, profiles, nil))

	byName := make(map[string]model.CodeMetrics)
	for _, node := range runtimeGraph.Nodes {
		byName[node.Static.Name] = node.Metrics
	}

	get := byName["GET"]
	require.NotNil(t, get)
	assert.Equal(t, int64(3), get[model.Metric{Source: trace.ProfileName, Name: "calls", Unit: "count"}].Self)
	assert.Equal(t, int64(9000), get[model.Metric{Source: trace.ProfileName, Name: "time", Unit: "nanoseconds"}].Self)

	set := byName["SET"]
	require.NotNil(t, set)
	assert.Equal(t, int64(5), set[model.Metric{Source: trace.ProfileName, Name: "calls", Unit: "count"}].Self)
}
