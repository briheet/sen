package process

import (
	"context"
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
	_ "net"
	_ "os"
	_ "runtime"
	_ "runtime/debug"
	_ "runtime/metrics"
	_ "runtime/pprof"
	_ "runtime/trace"
	_ "time"
)

func init() {}
`
)

// CreateOverlay writes the collector and its Go build overlay.
func CreateOverlay(_ context.Context, sourceDir string, tempDir string) (string, error) {
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
