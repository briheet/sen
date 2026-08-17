package analysis

import (
	"encoding/json"
	"fmt"
	"testing"
)

func benchGraph() []byte {
	var helper helperGraph
	functions := 2400
	for i := 0; i < functions; i++ {
		helper.Functions = append(helper.Functions, fmt.Sprintf("fn%d", i))
	}
	for i := 1; i < functions; i++ {
		helper.Edges = append(helper.Edges, helperEdge{From: uint64(i - 1), To: uint64(i)})
	}
	data, err := json.Marshal(helper)
	if err != nil {
		panic(err)
	}
	return data
}

func BenchmarkParse(b *testing.B) {
	data := benchGraph()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := parse(data, "/proj/calc.ml"); err != nil {
			b.Fatal(err)
		}
	}
}
