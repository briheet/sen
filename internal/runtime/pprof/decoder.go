package pprof

import (
	"io"
	"time"

	googleprofile "github.com/google/pprof/profile"
)

// Read decodes a Go pprof profile into Senbon's representation.
func Read(r io.Reader) (*Profile, error) {
	source, err := googleprofile.Parse(r)
	if err != nil {
		return nil, err
	}

	result := &Profile{
		Duration:          time.Duration(source.DurationNanos),
		DefaultSampleType: source.DefaultSampleType,
		Period:            source.Period,
		Locations:         make(map[LocationID]Location, len(source.Location)),
	}
	if source.TimeNanos != 0 {
		result.StartedAt = time.Unix(0, source.TimeNanos)
	}
	if source.PeriodType != nil {
		result.PeriodType = ValueType{Type: source.PeriodType.Type, Unit: source.PeriodType.Unit}
	}
	for _, sampleType := range source.SampleType {
		result.SampleTypes = append(result.SampleTypes, ValueType{Type: sampleType.Type, Unit: sampleType.Unit})
	}
	for _, sourceLocation := range source.Location {
		location := Location{ID: LocationID(sourceLocation.ID), Address: sourceLocation.Address}
		for _, line := range sourceLocation.Line {
			frame := Frame{Line: line.Line, Column: line.Column}
			if line.Function != nil {
				frame.Function = line.Function.Name
				frame.File = line.Function.Filename
				frame.StartLine = line.Function.StartLine
			}
			location.Frames = append(location.Frames, frame)
		}
		result.Locations[location.ID] = location
	}
	for _, sourceSample := range source.Sample {
		sample := Sample{Values: append([]int64(nil), sourceSample.Value...)}
		for _, location := range sourceSample.Location {
			sample.Stack = append(sample.Stack, LocationID(location.ID))
		}
		result.Samples = append(result.Samples, sample)
	}
	return result, nil
}
