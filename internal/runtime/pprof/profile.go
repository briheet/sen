// Package pprof defines Senbon's representation of Go pprof profiles.
package pprof

import "time"

// LocationID identifies a deduplicated profile location.
type LocationID uint64

// Profile contains sampled stacks and their measurement types.
type Profile struct {
	StartedAt         time.Time
	Duration          time.Duration
	SampleTypes       []ValueType
	DefaultSampleType string
	PeriodType        ValueType
	Period            int64
	Samples           []Sample
	Locations         map[LocationID]Location
}

// ValueType describes the meaning and unit of a sample value.
type ValueType struct {
	Type string
	Unit string
}

// Sample contains values corresponding to Profile.SampleTypes and a leaf-first stack.
type Sample struct {
	Values []int64
	Stack  []LocationID
}

// Location represents one address and its inlined call frames.
type Location struct {
	ID      LocationID
	Address uint64
	Frames  []Frame
}

// Frame identifies one source location in a sampled call stack.
type Frame struct {
	Function  string
	File      string
	Line      int64
	Column    int64
	StartLine int64
}
