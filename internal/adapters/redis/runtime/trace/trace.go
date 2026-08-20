// Package trace converts Redis commandstats counters into source-like profiles.
//
// Redis has no stack trace for a command. Sen models each known command as one
// synthetic frame and uses command call/time deltas as profile sample values.
package trace

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briheet/sen/internal/adapters/redis/analysis"
	"github.com/briheet/sen/internal/model"
)

const (
	// ProfileName identifies Redis command activity in the runtime graph.
	ProfileName = "redis"
	unitCount   = "count"
	unitNsec    = "nanoseconds"
)

// Counters contains the cumulative values Redis reports for one command.
type Counters struct {
	Calls        uint64
	Microseconds uint64
}

// Snapshot contains cumulative command counters keyed by uppercase command.
type Snapshot map[string]Counters

// Parse reads cmdstat_* rows from an INFO response. Unknown commands are
// ignored because the synthetic graph cannot attribute them to a node.
func Parse(info string) Snapshot {
	result := make(Snapshot)
	for line := range strings.SplitSeq(info, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if !strings.HasPrefix(line, "cmdstat_") {
			continue
		}
		name, fields, ok := strings.Cut(strings.TrimPrefix(line, "cmdstat_"), ":")
		name = strings.ToUpper(name)
		if !ok || !analysis.IsKnownCommand(name) {
			continue
		}
		result[name] = parseCounters(fields)
	}
	return result
}

func parseCounters(fields string) Counters {
	var result Counters
	for field := range strings.SplitSeq(fields, ",") {
		name, raw, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			continue
		}
		switch name {
		case "calls":
			result.Calls = value
		case "usec":
			result.Microseconds = value
		}
	}
	return result
}

// Delta returns the activity since previous. A missing command or a smaller
// counter is treated as new activity after a Redis restart.
func (current Snapshot) Delta(previous Snapshot) Snapshot {
	result := make(Snapshot, len(current))
	for name, counters := range current {
		before, ok := previous[name]
		if ok && counters.Calls >= before.Calls && counters.Microseconds >= before.Microseconds {
			counters.Calls -= before.Calls
			counters.Microseconds -= before.Microseconds
		}
		if counters.Calls > 0 || counters.Microseconds > 0 {
			result[name] = counters
		}
	}
	return result
}

// Profile maps one command window onto the synthetic Redis source graph.
func (snapshot Snapshot) Profile(duration time.Duration) *model.Profile {
	names := make([]string, 0, len(snapshot))
	for name := range snapshot {
		names = append(names, name)
	}
	sort.Strings(names)

	result := &model.Profile{
		Duration:    duration,
		SampleTypes: []model.ValueType{{Type: "calls", Unit: unitCount}, {Type: "time", Unit: unitNsec}},
		Locations:   make(map[model.ProfileLocationID]model.ProfileLocation, len(names)),
		Samples:     make([]model.ProfileSample, 0, len(names)),
	}
	for index, name := range names {
		locationID := model.ProfileLocationID(index + 1)
		result.Locations[locationID] = model.ProfileLocation{
			ID: locationID,
			Frames: []model.ProfileFrame{{
				Function: name,
				File:     analysis.ModulePath + "/" + name,
				Line:     1,
			}},
		}
		counters := snapshot[name]
		result.Samples = append(result.Samples, model.ProfileSample{
			Values: []int64{saturatingScale(counters.Calls, 1), saturatingScale(counters.Microseconds, 1000)},
			Stack:  []model.ProfileLocationID{locationID},
		})
	}
	return result
}

func saturatingScale(value, scale uint64) int64 {
	if value > uint64(math.MaxInt64)/scale {
		return math.MaxInt64
	}
	return int64(value * scale)
}
