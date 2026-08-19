package golang

import (
	"context"
	"testing"

	"github.com/briheet/sen/internal/adapters/golang/analysis"
)

// BenchmarkAnalysisPhases keeps package loading and graph projection measurable
// independently so startup regressions are visible.
func BenchmarkAnalysisPhases(b *testing.B) {
	const source = "../../../examples/go/http"
	b.Run("load", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := LoadPackages(context.Background(), source, nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	packages, err := LoadPackages(context.Background(), source, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("graph", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := analysis.GetGraph(packages); err != nil {
				b.Fatal(err)
			}
		}
	})
}
