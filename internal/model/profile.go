package model

import "time"

// ProfileLocationID identifies a deduplicated profile location.
type ProfileLocationID uint64

// Profile contains sampled stacks and their measurement types.
type Profile struct {
	StartedAt         time.Time
	Duration          time.Duration
	SampleTypes       []ValueType
	DefaultSampleType string
	PeriodType        ValueType
	Period            int64
	Samples           []ProfileSample
	Locations         map[ProfileLocationID]ProfileLocation
}

// ValueType describes the meaning and unit of a sample value.
type ValueType struct {
	Type string
	Unit string
}

// ProfileSample contains values and a leaf-first stack.
type ProfileSample struct {
	Values []int64
	Stack  []ProfileLocationID
}

// ProfileLocation represents one address and its inlined frames.
type ProfileLocation struct {
	ID      ProfileLocationID
	Address uint64
	Frames  []ProfileFrame
}

// ProfileFrame identifies one sampled source location.
type ProfileFrame struct {
	Function  string
	File      string
	Line      int64
	Column    int64
	StartLine int64
}
