package metrics

import "testing"

func BenchmarkDecode(b *testing.B) {
	db := Database{}
	b.ReportAllocs()
	for b.Loop() {
		m := Decode(db)
		_ = m
	}
}
