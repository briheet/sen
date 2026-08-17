package runtime

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
	VirtualFilePrefix    = "zz_"
	VirtualFileExtension = ".go"
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

func CreateOverlay(ctx context.Context, sourceDir string, tempDir string) (string, error) {
	// Absolute path + Collector path for overlay adding
	collectorPath := filepath.Join(sourceDir, CollectorFileName)
	if err := os.WriteFile(collectorPath, []byte(CollectorSource), 0o600); err != nil {
		return "", err
	}

	// Create a virtual filepath for overlay json addition
	virtualName := virtualFilePrefix + strings.ReplaceAll(filepath.Base(tempDir), "-", "_") + virtualFileExtension

	// Overlay json file
	// {
	//   "Replace": {
	//     "/code/myapp/zz_senbon_12345.go": "/tmp/senbon-12345/collector.go"
	//   }
	// }
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

	// Join overlay file to tempDir and write over the template
	overlayPath := filepath.Join(tempDir, OverlayFileName)
	if err = os.WriteFile(overlayPath, overlay, 0o600); err != nil {
		return "", err
	}

	return overlayPath, nil
}
