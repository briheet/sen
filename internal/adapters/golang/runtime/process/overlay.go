package process

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	CollectorFileName    = "collector.go"
	OverlayFileName      = "overlay.json"
	virtualFilePrefix    = "zz_"
	virtualFileExtension = ".go"
	CollectorSource      = `package main

import (
	"encoding/gob"
	"net"
	"net/http"
	runtimepprof "net/http/pprof"
	"os"
	runtimemetrics "runtime/metrics"
)

const senbonCollectorSocket = "SENBON_COLLECTOR_SOCKET"

type senbonCollectorMetric struct {
	Name      string
	Uint64    uint64
	Float64   float64
	Histogram *runtimemetrics.Float64Histogram
}

func init() {
	socket := os.Getenv(senbonCollectorSocket)
	if socket == "" {
		return
	}
	_ = os.Remove(socket)
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/senbon/metrics", senbonCollectorMetrics)
	mux.HandleFunc("/debug/pprof/", runtimepprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", runtimepprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", runtimepprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", runtimepprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", runtimepprof.Trace)
	go func() { _ = http.Serve(listener, mux) }()
}

func senbonCollectorMetrics(writer http.ResponseWriter, _ *http.Request) {
	descriptions := runtimemetrics.All()
	samples := make([]runtimemetrics.Sample, len(descriptions))
	for index, description := range descriptions {
		samples[index].Name = description.Name
	}
	runtimemetrics.Read(samples)

	result := make([]senbonCollectorMetric, 0, len(samples))
	for _, sample := range samples {
		metric := senbonCollectorMetric{Name: sample.Name}
		switch sample.Value.Kind() {
		case runtimemetrics.KindUint64:
			metric.Uint64 = sample.Value.Uint64()
		case runtimemetrics.KindFloat64:
			metric.Float64 = sample.Value.Float64()
		case runtimemetrics.KindFloat64Histogram:
			metric.Histogram = sample.Value.Float64Histogram()
		default:
			continue
		}
		result = append(result, metric)
	}
	writer.Header().Set("Content-Type", "application/octet-stream")
	_ = gob.NewEncoder(writer).Encode(result)
}
`
)

// CreateOverlay writes the collector and its Go build overlay.
func CreateOverlay(sourceDir string, tempDir string) (string, error) {
	collectorPath := filepath.Join(tempDir, CollectorFileName)
	if err := os.WriteFile(collectorPath, []byte(CollectorSource), 0o600); err != nil {
		return "", err
	}

	virtualName := virtualFilePrefix + strings.ReplaceAll(filepath.Base(tempDir), "-", "_") + virtualFileExtension
	overlay, err := json.Marshal(struct {
		Replace map[string]string
	}{
		map[string]string{
			filepath.Join(sourceDir, virtualName): collectorPath,
		},
	})
	if err != nil {
		return "", err
	}

	overlayPath := filepath.Join(tempDir, OverlayFileName)
	if err = os.WriteFile(overlayPath, overlay, 0o600); err != nil {
		return "", err
	}
	return overlayPath, nil
}
