// Package metrics decodes TigerBeetle's DogStatsD wire format.
package metrics

import (
	"errors"
	"strconv"
	"strings"
)

// Sample is one validated tb.* metric line.
type Sample struct {
	Name      string
	Value     float64
	Type      string
	Cluster   string
	Replica   uint32
	Operation string
	Tree      string
}

// ParseLine decodes a gauge or counter with required cluster and replica tags.
func ParseLine(line string) (Sample, error) {
	line = strings.TrimSpace(line)
	nameValue, fields, ok := strings.Cut(line, "|")
	if !ok {
		return Sample{}, errors.New("missing metric type")
	}
	name, valueText, ok := strings.Cut(nameValue, ":")
	if !ok || !strings.HasPrefix(name, "tb.") {
		return Sample{}, errors.New("not a TigerBeetle metric")
	}
	parts := strings.Split(fields, "|")
	if len(parts) < 2 || parts[0] != "g" && parts[0] != "c" || !strings.HasPrefix(parts[1], "#") {
		return Sample{}, errors.New("unsupported DogStatsD fields")
	}
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil || value < 0 {
		return Sample{}, errors.New("invalid metric value")
	}
	sample := Sample{Name: strings.TrimPrefix(name, "tb."), Value: value, Type: parts[0]}
	var replicaText string
	for _, tag := range strings.Split(strings.TrimPrefix(parts[1], "#"), ",") {
		key, tagValue, found := strings.Cut(tag, ":")
		if !found {
			continue
		}
		switch key {
		case "cluster":
			sample.Cluster = tagValue
		case "replica":
			replicaText = tagValue
		case "operation":
			sample.Operation = tagValue
		case "tree":
			sample.Tree = tagValue
		}
	}
	if !validCluster(sample.Cluster) || replicaText == "" {
		return Sample{}, errors.New("missing cluster or replica tag")
	}
	replica, err := strconv.ParseUint(replicaText, 10, 32)
	if err != nil {
		return Sample{}, errors.New("invalid replica tag")
	}
	sample.Replica = uint32(replica)
	return sample, nil
}

func validCluster(cluster string) bool {
	if len(cluster) != 32 {
		return false
	}
	for _, character := range cluster {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return true
}
