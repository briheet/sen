//go:build linux

package profiler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/briheet/sen/internal/adapters/rust/analysis"
	"github.com/briheet/sen/internal/model"
)

var perfFrame = regexp.MustCompile(`^\s*([[:xdigit:]]+)\s+(.+?)(?:\s+\([^)]*\))?(?:\s+(.+:[0-9]+))?$`)

// Capture records one strict one-second perf window.
func Capture(ctx context.Context, pid int, symbols *analysis.Symbols) (*model.Profile, *model.Trace, uint64, error) {
	path, err := temporaryProfile("sen-rust-perf-*")
	if err != nil {
		return nil, nil, 0, err
	}
	defer os.Remove(path)
	record := exec.CommandContext(ctx, "perf", "record", "-q", "-F", "99", "-g", "--call-graph", "fp", "-p", strconv.Itoa(pid), "-o", path, "--", "sleep", "1")
	if output, err := record.CombinedOutput(); err != nil {
		return nil, nil, 0, fmt.Errorf("perf record Rust process: %w: %s", err, strings.TrimSpace(string(output)))
	}
	script := exec.CommandContext(ctx, "perf", "script", "-i", path, "-F", "ip,sym,dso,srcline")
	output, err := script.Output()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("perf script Rust process: %w", err)
	}
	samples := parsePerf(string(output))
	profile, trace, count := normalize(samples, symbols)
	if count == 0 {
		return nil, nil, 0, fmt.Errorf("perf returned no Rust stacks")
	}
	return profile, trace, count, nil
}

func parsePerf(output string) []rawSample {
	var result []rawSample
	var frames []rawFrame
	flush := func() {
		if len(frames) > 0 {
			result = append(result, rawSample{frames: frames, weight: 1})
			frames = nil
		}
	}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		match := perfFrame.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		address, _ := strconv.ParseUint(match[1], 16, 64)
		file, lineNumber := sourceLocation(match[3])
		frames = append(frames, rawFrame{address: address, function: strings.TrimSpace(match[2]), file: file, line: lineNumber})
	}
	flush()
	return result
}

func temporaryProfile(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	return path, file.Close()
}
