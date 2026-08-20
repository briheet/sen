package metrics

import "testing"

func BenchmarkDecode(b *testing.B) {
	body := "# Memory\r\n" +
		"used_memory:1048576\r\nused_memory_dataset:524288\r\nused_memory_rss:2097152\r\n" +
		"# Stats\r\ntotal_commands_processed:300\ntotal_connections_received:12\r\n" +
		"# CPU\r\nused_cpu_user:0.5\r\nused_cpu_sys:0.25\r\n" +
		"# Clients\r\nconnected_clients:4\r\n"
	b.ReportAllocs()
	for b.Loop() {
		m := Decode(body)
		_ = m
	}
}
