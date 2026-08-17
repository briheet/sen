package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkReadMetrics(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "metrics.ndjson")
	file, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	line := []byte(`{"heapUsed":1000,"heapTotal":2000,"rss":3000,"user":1,"system":2}` + "\n")
	for index := 0; index < 1000; index++ {
		if _, err := file.Write(line); err != nil {
			b.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		offset := int64(0)
		if _, ok := readMetrics(path, &offset); !ok {
			b.Fatal("no metrics")
		}
	}
}
