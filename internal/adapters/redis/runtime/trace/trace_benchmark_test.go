package trace

import "testing"

func BenchmarkDecode(b *testing.B) {
	body := ""
	for i := 0; i < 50; i++ {
		body += "cmdstat_get:calls=100,usec=900,usec_per_call=9.00,rejected_calls=0,failed_calls=0\r\n" +
			"cmdstat_set:calls=80,usec=240,usec_per_call=3.00,rejected_calls=0,failed_calls=0\r\n"
	}
	b.ReportAllocs()
	for b.Loop() {
		p := Decode(body)
		_ = p
	}
}
