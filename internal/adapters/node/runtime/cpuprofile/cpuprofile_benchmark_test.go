package cpuprofile

import (
	"encoding/json"
	"strconv"
	"testing"
)

// benchProfile builds a binary-tree call tree with one second of samples.
func benchProfile() *CPUProfile {
	const nodes = 400
	profile := &CPUProfile{EndTime: 1000 * 1000}
	for id := 1; id <= nodes; id++ {
		node := Node{
			ID: uint32(id),
			CallFrame: CallFrame{
				FunctionName: "fn" + strconv.Itoa(id),
				URL:          "/app/src/index.js",
				LineNumber:   id,
				ColumnNumber: id % 40,
			},
		}
		if left := 2 * id; left <= nodes {
			node.Children = []uint32{uint32(left), uint32(left + 1)}
		}
		profile.Nodes = append(profile.Nodes, node)
	}
	for index := 0; index < 1000; index++ {
		profile.Samples = append(profile.Samples, uint32(index%nodes+1))
		profile.TimeDeltas = append(profile.TimeDeltas, 1000)
	}
	return profile
}

func BenchmarkParse(b *testing.B) {
	raw, err := json.Marshal(benchProfile())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var profile CPUProfile
		if err := json.Unmarshal(raw, &profile); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProfile(b *testing.B) {
	profile := benchProfile()
	b.ReportAllocs()
	for b.Loop() {
		if result := profile.Profile(); len(result.Samples) == 0 {
			b.Fatal("no samples")
		}
	}
}

func BenchmarkTrace(b *testing.B) {
	profile := benchProfile()
	b.ReportAllocs()
	for b.Loop() {
		if result := profile.Trace(); len(result.Events) == 0 {
			b.Fatal("no events")
		}
	}
}

func BenchmarkDecodePipeline(b *testing.B) {
	raw, err := json.Marshal(benchProfile())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var profile CPUProfile
		if err := json.Unmarshal(raw, &profile); err != nil {
			b.Fatal(err)
		}
		if len(profile.Profile().Samples) == 0 {
			b.Fatal("no samples")
		}
		if len(profile.Trace().Events) == 0 {
			b.Fatal("no events")
		}
	}
}
