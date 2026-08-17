package runtime

import (
	"encoding/binary"
	"testing"
)

// benchEvents builds a synthetic runtime events ring with a metadata header and
// a run of GC exit events.
func benchEvents(count int) []byte {
	meta := make([]byte, 64)
	binary.LittleEndian.PutUint64(meta[0:], 1)    // version
	binary.LittleEndian.PutUint64(meta[6*8:], 64) // data offset
	total := append([]byte(nil), meta...)
	for i := 0; i < count; i++ {
		var w uint64
		w |= 1 << 54          // length=1 (has payload word)
		w |= uint64(3) << 49  // type = EXIT
		w |= uint64(17) << 36 // id = MINOR
		total = binary.LittleEndian.AppendUint64(total, w)
		total = binary.LittleEndian.AppendUint64(total, 0) // payload word
	}
	return total
}

func BenchmarkDecodeEvents(b *testing.B) {
	buf := benchEvents(1000)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := decodeEvents(buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeEventsSparse mimics a real 68MB ring: a short run of real
// items at the start of an otherwise sparse (zero-padded) buffer.
func BenchmarkDecodeEventsSparse(b *testing.B) {
	buf := make([]byte, 68*1024*1024)
	copy(buf, benchEvents(1000))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := decodeEvents(buf); err != nil {
			b.Fatal(err)
		}
	}
}
