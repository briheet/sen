// Package trace turns cumulative PostgreSQL statistics into windowed profiles.
package trace

import (
	"math"
	"sort"
	"time"

	"github.com/briheet/sen/internal/adapters/postgres/analysis"
	"github.com/briheet/sen/internal/model"
)

const (
	// StatementsSource identifies activity from pg_stat_statements.
	StatementsSource = "statements"
	// TablesSource identifies activity from pg_stat_user_tables.
	TablesSource = "tables"
)

// Statement is one cumulative pg_stat_statements row.
type Statement struct {
	QueryID    int64
	Query      string
	Label      string
	Calls      int64
	TotalExec  float64 // milliseconds
	Rows       int64
	BlocksRead int64
}

// Table is one cumulative pg_stat_user_tables row.
type Table struct {
	Name       string
	SeqScan    int64
	IdxScan    int64
	Inserts    int64
	Updates    int64
	Deletes    int64
	LiveTuples int64
}

// Snapshot indexes one collection so cumulative counters can be differenced.
type Snapshot struct {
	statements map[int64]Statement
	tables     map[string]Table
}

// NewSnapshot indexes the rows returned by one collection.
func NewSnapshot(statements []Statement, tables []Table) Snapshot {
	result := Snapshot{
		statements: make(map[int64]Statement, len(statements)),
		tables:     make(map[string]Table, len(tables)),
	}
	for _, statement := range statements {
		if statement.Label == "" {
			statement.Label = analysis.StatementLabel(statement.Query)
		}
		result.statements[statement.QueryID] = statement
	}
	for _, table := range tables {
		result.tables[table.Name] = table
	}
	return result
}

// Initialized reports whether the snapshot came from a collection.
func (s Snapshot) Initialized() bool { return s.statements != nil || s.tables != nil }

// Delta returns activity since previous. Smaller counters are treated as a
// statistics reset and the current value becomes the new window.
func (s Snapshot) Delta(previous Snapshot) Snapshot {
	result := NewSnapshot(nil, nil)
	for id, current := range s.statements {
		before, found := previous.statements[id]
		if found {
			current.Calls = deltaInt(current.Calls, before.Calls)
			current.TotalExec = deltaFloat(current.TotalExec, before.TotalExec)
			current.Rows = deltaInt(current.Rows, before.Rows)
			current.BlocksRead = deltaInt(current.BlocksRead, before.BlocksRead)
		}
		if current.Calls > 0 || current.TotalExec > 0 {
			result.statements[id] = current
		}
	}
	for name, current := range s.tables {
		before, found := previous.tables[name]
		if found {
			current.SeqScan = deltaInt(current.SeqScan, before.SeqScan)
			current.IdxScan = deltaInt(current.IdxScan, before.IdxScan)
			current.Inserts = deltaInt(current.Inserts, before.Inserts)
			current.Updates = deltaInt(current.Updates, before.Updates)
			current.Deletes = deltaInt(current.Deletes, before.Deletes)
		}
		if tableOperations(current) > 0 {
			result.tables[name] = current
		}
	}
	return result
}

// Profiles maps a window onto the synthetic statement and table graph.
func (s Snapshot) Profiles(duration time.Duration) map[string]*model.Profile {
	result := make(map[string]*model.Profile, 2)
	if len(s.statements) > 0 {
		result[StatementsSource] = s.statementProfile(duration)
	}
	if len(s.tables) > 0 {
		result[TablesSource] = s.tableProfile(duration)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func (s Snapshot) statementProfile(duration time.Duration) *model.Profile {
	ids := make([]int64, 0, len(s.statements))
	for id := range s.statements {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	profile := newProfile(duration, len(ids),
		[]model.ValueType{{Type: "calls", Unit: "count"}, {Type: "time", Unit: "nanoseconds"}},
	)
	for index, id := range ids {
		statement := s.statements[id]
		location := model.ProfileLocationID(index + 1)
		profile.Locations[location] = model.ProfileLocation{ID: location, Frames: []model.ProfileFrame{{
			Function: statement.Label, File: analysis.StmtPath(id), Line: 1,
		}}}
		profile.Samples = append(profile.Samples, model.ProfileSample{
			Values: []int64{max(0, statement.Calls), millisecondsToNanoseconds(statement.TotalExec)},
			Stack:  []model.ProfileLocationID{location},
		})
	}
	return profile
}

func (s Snapshot) tableProfile(duration time.Duration) *model.Profile {
	names := make([]string, 0, len(s.tables))
	for name := range s.tables {
		names = append(names, name)
	}
	sort.Strings(names)
	profile := newProfile(duration, len(names), []model.ValueType{{Type: "operations", Unit: "count"}})
	for index, name := range names {
		location := model.ProfileLocationID(index + 1)
		profile.Locations[location] = model.ProfileLocation{ID: location, Frames: []model.ProfileFrame{{
			Function: name, File: analysis.TablePath(name), Line: 1,
		}}}
		profile.Samples = append(profile.Samples, model.ProfileSample{
			Values: []int64{tableOperations(s.tables[name])},
			Stack:  []model.ProfileLocationID{location},
		})
	}
	return profile
}

func newProfile(duration time.Duration, size int, types []model.ValueType) *model.Profile {
	return &model.Profile{
		Duration: duration, SampleTypes: types,
		Locations: make(map[model.ProfileLocationID]model.ProfileLocation, size),
		Samples:   make([]model.ProfileSample, 0, size),
	}
}

func deltaInt(current, previous int64) int64 {
	if current >= previous {
		return current - previous
	}
	return max(0, current)
}

func deltaFloat(current, previous float64) float64 {
	if current >= previous {
		return current - previous
	}
	return max(0, current)
}

func tableOperations(table Table) int64 {
	return max(0, table.SeqScan) + max(0, table.IdxScan) + max(0, table.Inserts) +
		max(0, table.Updates) + max(0, table.Deletes)
}

func millisecondsToNanoseconds(value float64) int64 {
	value *= float64(time.Millisecond)
	if value >= math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(max(0, value))
}
