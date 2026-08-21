package cdp

import (
	"bufio"
	"encoding/json"
	"net"
	"strconv"
	"testing"

	"github.com/briheet/sen/internal/adapters/node/runtime/cpuprofile"
)

const benchPayloadSize = 1024

func BenchmarkWriteFrame(b *testing.B) {
	server, client := net.Pipe()
	defer func() { _ = server.Close() }()
	defer func() { _ = client.Close() }()
	writer := &Client{conn: client}
	reader := &Client{conn: server, reader: bufio.NewReader(server)}
	go func() {
		for {
			if _, err := reader.readFrame(); err != nil {
				return
			}
		}
	}()

	payload := make([]byte, benchPayloadSize)
	b.ReportAllocs()
	for b.Loop() {
		if err := writer.writeText(payload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadFrame(b *testing.B) {
	server, client := net.Pipe()
	defer func() { _ = server.Close() }()
	defer func() { _ = client.Close() }()
	writer := &Client{conn: server}
	reader := &Client{conn: client, reader: bufio.NewReader(client)}
	go func() {
		payload := make([]byte, benchPayloadSize)
		for {
			if err := writer.writeText(payload); err != nil {
				return
			}
		}
	}()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := reader.readFrame(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeHeader(b *testing.B) {
	payload := benchmarkProfilePayload(b)
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for b.Loop() {
		var header struct {
			ID uint64 `json:"id"`
		}
		if err := json.Unmarshal(payload, &header); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(len(payload)), "payload-B")
}

func BenchmarkDecodeResponse(b *testing.B) {
	b.Run("Profile", func(b *testing.B) {
		payload := benchmarkProfilePayload(b)
		b.ReportAllocs()
		b.SetBytes(int64(len(payload)))
		for b.Loop() {
			var response struct {
				Profile cpuprofile.CPUProfile `json:"profile"`
			}
			if err := decodeResponse("Profiler.stop", payload, &response); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(len(payload)), "payload-B")
	})

	b.Run("Metrics", func(b *testing.B) {
		payload, err := json.Marshal(struct {
			ID     uint64 `json:"id"`
			Result any    `json:"result"`
		}{ID: 1, Result: map[string]any{
			"result": map[string]any{"value": benchmarkMetrics{}},
		}})
		if err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		b.SetBytes(int64(len(payload)))
		for b.Loop() {
			var response struct {
				Result struct {
					Value benchmarkMetrics `json:"value"`
				} `json:"result"`
			}
			if err := decodeResponse("Runtime.evaluate", payload, &response); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(len(payload)), "payload-B")
	})
}

type benchmarkMetrics struct {
	HeapUsed             uint64  `json:"heapUsed"`
	HeapTotal            uint64  `json:"heapTotal"`
	External             uint64  `json:"external"`
	ArrayBuffers         uint64  `json:"arrayBuffers"`
	EventLoopUtilization float64 `json:"eventLoopUtilization"`
	EventLoopDelayMean   float64 `json:"eventLoopDelayMean"`
	EventLoopDelayMax    float64 `json:"eventLoopDelayMax"`
	EventLoopDelayP95    float64 `json:"eventLoopDelayP95"`
	EventLoopDelayP99    float64 `json:"eventLoopDelayP99"`
	ActiveResources      uint64  `json:"activeResources"`
}

func benchmarkProfilePayload(tb testing.TB) []byte {
	tb.Helper()
	const nodeCount = 400
	profile := cpuprofile.CPUProfile{EndTime: 1000 * 1000}
	for id := 1; id <= nodeCount; id++ {
		node := cpuprofile.Node{
			ID: uint32(id),
			CallFrame: cpuprofile.CallFrame{
				FunctionName: "fn" + strconv.Itoa(id),
				URL:          "/app/src/index.js",
				LineNumber:   id,
				ColumnNumber: id % 40,
			},
		}
		if left := 2 * id; left <= nodeCount {
			node.Children = []uint32{uint32(left), uint32(left + 1)}
		}
		profile.Nodes = append(profile.Nodes, node)
	}
	for index := range 1000 {
		profile.Samples = append(profile.Samples, uint32(index%nodeCount+1))
		profile.TimeDeltas = append(profile.TimeDeltas, 1000)
	}
	payload, err := json.Marshal(struct {
		ID     uint64 `json:"id"`
		Result any    `json:"result"`
	}{ID: 1, Result: struct {
		Profile cpuprofile.CPUProfile `json:"profile"`
	}{Profile: profile}})
	if err != nil {
		tb.Fatal(err)
	}
	return payload
}
