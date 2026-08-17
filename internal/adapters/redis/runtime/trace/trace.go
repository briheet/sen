// Package trace builds per-command heat and latency profiles for Redis.
//
// Redis is observed over its public protocol: INFO commandstats/latencystats
// provide cumulative per-command call counts and execution time, while
// SLOWLOG GET surfaces the individual slow command executions. None of these
// require injecting anything into the server, so tracing is reduced to parsing
// the text replies and folding them into a normalized profile.
package trace

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briheet/sen/internal/adapters/redis/analysis"
	"github.com/briheet/sen/internal/model"
)

const (
	// ProfileName is the profile source name used for per-command heat.
	ProfileName = "redis"
	unitCount   = "count"
	unitNsec    = "nanoseconds"
)

// commandHeat accumulates per-command observed totals.
type commandHeat struct {
	calls uint64
	usec  uint64
}

// profile maps a command name to its accumulated heat.
type profile struct {
	heat map[string]*commandHeat
}

func newProfile() *profile {
	return &profile{heat: make(map[string]*commandHeat)}
}

// addCommand records observed calls/useconds for a command.
func (p *profile) addCommand(name string, calls, usec uint64) {
	entry := p.heat[name]
	if entry == nil {
		entry = &commandHeat{}
		p.heat[name] = entry
	}
	entry.calls += calls
	entry.usec += usec
}

// Decode builds a normalized per-command profile from a commandstats body and
// the parsed status fields of the server. The body should be the output of
// `INFO commandstats`.
func Decode(cmdstats string) *model.Profile {
	p := newProfile()
	for _, line := range strings.Split(cmdstats, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if !strings.HasPrefix(line, "cmdstat_") {
			continue
		}
		rest := strings.TrimPrefix(line, "cmdstat_")
		name, stats, ok := strings.Cut(rest, ":")
		if !ok {
			continue
		}
		name = strings.ToUpper(name)
		calls, usec := parseCmdstat(stats)
		if !analysis.IsKnownCommand(name) {
			continue
		}
		p.addCommand(name, calls, usec)
	}
	return p.profile()
}

// parseCmdstat extracts calls and usec from a "calls=1,usec=2,..." fragment.
func parseCmdstat(stats string) (uint64, uint64) {
	var calls, usec uint64
	for stats != "" {
		var field string
		if rest, after, ok := strings.Cut(stats, ","); ok {
			field, stats = rest, after
		} else {
			field, stats = stats, ""
		}
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		switch key {
		case "calls":
			calls = parseUint(value)
		case "usec":
			usec = parseUint(value)
		}
	}
	return calls, usec
}

func parseUint(value string) uint64 {
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// profile folds accumulated heat into a model.Profile whose samples point at
// the synthetic command nodes, so the runtime mapper can attribute them.
func (p *profile) profile() *model.Profile {
	names := make([]string, 0, len(p.heat))
	for name := range p.heat {
		names = append(names, name)
	}
	sort.Strings(names)

	result := &model.Profile{
		Duration:    time.Second,
		SampleTypes: []model.ValueType{{Type: "calls", Unit: unitCount}, {Type: "usec", Unit: unitNsec}},
		Locations:   make(map[model.ProfileLocationID]model.ProfileLocation, len(names)),
	}
	if len(names) == 0 {
		return result
	}

	var id model.ProfileLocationID = 1
	for _, name := range names {
		entry := p.heat[name]
		result.Locations[id] = model.ProfileLocation{
			ID: id,
			Frames: []model.ProfileFrame{{
				Function: name,
				File:     analysis.ModulePath + "/" + name,
				Line:     1,
			}},
		}
		result.Samples = append(result.Samples, model.ProfileSample{
			Values: []int64{int64(entry.calls), int64(entry.usec)},
			Stack:  []model.ProfileLocationID{id},
		})
		id++
	}
	return result
}
