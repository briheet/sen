package analysis

import (
	"encoding/json"
	"fmt"
	"testing"
)

// benchGraph builds helper output for 40 files with 60 functions each.
func benchGraph() []byte {
	var helper helperGraph
	nextID := 0
	for fileIndex := 0; fileIndex < 40; fileIndex++ {
		file := helperFile{Path: fmt.Sprintf("/app/src/file%d.js", fileIndex)}
		for index := 0; index < 60; index++ {
			file.Functions = append(file.Functions, helperFunction{
				ID:        uint64(nextID),
				Name:      fmt.Sprintf("fn%d_%d", fileIndex, index),
				StartLine: index * 10,
				EndLine:   index*10 + 8,
			})
			nextID++
		}
		helper.Files = append(helper.Files, file)
	}
	for index := 1; index < nextID; index++ {
		helper.Edges = append(helper.Edges, helperEdge{From: uint64(index - 1), To: uint64(index)})
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
		if _, err := parse(data, "/app"); err != nil {
			b.Fatal(err)
		}
	}
}
