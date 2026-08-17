// Package trace builds per-query and per-table heat profiles for PostgreSQL.
// Rows come from pg_stat_statements and pg_stat_user_tables; each is folded
// into a normalized profile attributed to the matching synthetic graph node.
package trace

import (
	"strings"
	"time"

	"github.com/briheet/sen/internal/adapters/postgres/analysis"
	"github.com/briheet/sen/internal/model"
)

const (
	// StatementsSource is the profile source name for per-query heat.
	StatementsSource = "statements"
	// TablesSource is the profile source name for per-table heat.
	TablesSource = "tables"
	unitCount    = "count"
	unitMillis   = "milliseconds"
	maxLabel     = 60
)

// whitespace normalizes whitespace in statement text for single-line labels.
var whitespace = strings.NewReplacer("\n", " ", "\t", " ", "\r", " ")

// labelSampleTypes is the sample value schema shared by every profile built
// here; hoisted so it is not reallocated on each call.
var labelSampleTypes = []model.ValueType{{Type: "calls", Unit: unitCount}, {Type: "time", Unit: unitMillis}}

// Statement is one row sampled from pg_stat_statements.
type Statement struct {
	QueryID    int64
	Query      string
	Calls      int64
	TotalExec  float64 // milliseconds
	Rows       int64
	BlocksRead int64
}

// Table is one row sampled from pg_stat_user_tables.
type Table struct {
	Name       string
	SeqScan    int64
	IdxScan    int64
	Inserts    int64
	Updates    int64
	Deletes    int64
	LiveTuples int64
}

// Statements builds a normalized per-query profile keyed by query id path.
func Statements(rows []Statement) *model.Profile { return statements(rows) }

// Tables builds a normalized per-table profile keyed by table name path.
func Tables(rows []Table) *model.Profile { return tables(rows) }

// statements builds a normalized per-query profile keyed by query id path.
func statements(rows []Statement) *model.Profile {
	locations := make(map[model.ProfileLocationID]model.ProfileLocation, len(rows))
	samples := make([]model.ProfileSample, 0, len(rows))
	var id model.ProfileLocationID = 1
	for _, r := range rows {
		locations[id] = model.ProfileLocation{ID: id, Frames: []model.ProfileFrame{{
			Function: label(r.Query),
			File:     analysis.StmtPath(r.QueryID),
			Line:     1,
		}}}
		samples = append(samples, model.ProfileSample{
			Values: []int64{r.Calls, int64(r.TotalExec)},
			Stack:  []model.ProfileLocationID{id},
		})
		id++
	}
	return newProfile(StatementsSource, locations, samples)
}

// tables builds a normalized per-table profile keyed by table name path.
func tables(rows []Table) *model.Profile {
	locations := make(map[model.ProfileLocationID]model.ProfileLocation, len(rows))
	samples := make([]model.ProfileSample, 0, len(rows))
	var id model.ProfileLocationID = 1
	for _, r := range rows {
		locations[id] = model.ProfileLocation{ID: id, Frames: []model.ProfileFrame{{
			Function: r.Name,
			File:     analysis.TablePath(r.Name),
			Line:     1,
		}}}
		samples = append(samples, model.ProfileSample{
			Values: []int64{r.SeqScan + r.IdxScan, r.LiveTuples},
			Stack:  []model.ProfileLocationID{id},
		})
		id++
	}
	return newProfile(TablesSource, locations, samples)
}

func newProfile(source string, locations map[model.ProfileLocationID]model.ProfileLocation, samples []model.ProfileSample) *model.Profile {
	return &model.Profile{
		Duration:    time.Second,
		SampleTypes: labelSampleTypes,
		Locations:   locations,
		Samples:     samples,
	}
}

// label produces a short single-line label for a statement from its SQL text.
func label(query string) string {
	query = strings.TrimSpace(whitespace.Replace(query))
	if len(query) <= maxLabel {
		return query
	}
	// Only convert to runes when truncation is actually required.
	runes := []rune(query)
	if len(runes) <= maxLabel {
		return query
	}
	return string(runes[:maxLabel]) + "…"
}
