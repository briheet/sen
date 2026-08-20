//go:build darwin

package profiler

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/briheet/sen/internal/adapters/rust/analysis"
	"github.com/briheet/sen/internal/model"
)

var sampleLine = regexp.MustCompile(`^(\s*)(?:[+!?*|](\s*))?([0-9]+)\s+(.+)$`)

// Capture records one strict one-second sample window.
func Capture(ctx context.Context, pid int, symbols *analysis.Symbols) (*model.Profile, *model.Trace, uint64, error) {
	path, err := temporaryProfile("sen-rust-sample-*")
	if err != nil {
		return nil, nil, 0, err
	}
	defer os.Remove(path)
	cmd := exec.CommandContext(ctx, "/usr/bin/sample", strconv.Itoa(pid), "1", "10", "-mayDie", "-fullPaths", "-file", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, nil, 0, fmt.Errorf("sample Rust process: %w: %s", err, strings.TrimSpace(string(output)))
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer file.Close()
	samples, err := parseSample(file)
	if err != nil {
		return nil, nil, 0, err
	}
	profile, trace, count := normalize(samples, symbols)
	if count == 0 {
		return nil, nil, 0, fmt.Errorf("sample returned no Rust stacks")
	}
	return profile, trace, count, nil
}

type sampleNode struct {
	depth  int
	frame  rawFrame
	weight uint64
	child  bool
}

func parseSample(file io.Reader) ([]rawSample, error) {
	var result []rawSample
	var path []sampleNode
	inGraph := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "Call graph:" {
			inGraph = true
			continue
		}
		if !inGraph {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "Total number in stack") {
			break
		}
		match := sampleLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		weight, _ := strconv.ParseUint(match[3], 10, 64)
		depth := len(match[1]) + len(match[2])
		for len(path) > 0 && depth <= path[len(path)-1].depth {
			if !path[len(path)-1].child {
				result = append(result, pathSample(path))
			}
			path = path[:len(path)-1]
		}
		if len(path) > 0 {
			path[len(path)-1].child = true
		}
		path = append(path, sampleNode{depth: depth, frame: parseSampleFrame(match[4]), weight: weight})
	}
	for len(path) > 0 {
		if !path[len(path)-1].child {
			result = append(result, pathSample(path))
		}
		path = path[:len(path)-1]
	}
	return result, scanner.Err()
}

func pathSample(path []sampleNode) rawSample {
	leaf := path[len(path)-1]
	frames := make([]rawFrame, 0, len(path))
	for index := len(path) - 1; index >= 0; index-- {
		frames = append(frames, path[index].frame)
	}
	return rawSample{frames: frames, weight: leaf.weight}
}

func parseSampleFrame(value string) rawFrame {
	value = strings.TrimSpace(value)
	frame := rawFrame{}
	if fields := strings.Fields(value); len(fields) > 0 && strings.HasPrefix(fields[0], "0x") {
		frame.address, _ = strconv.ParseUint(strings.TrimPrefix(fields[0], "0x"), 16, 64)
		value = strings.TrimSpace(strings.TrimPrefix(value, fields[0]))
	}
	if index := strings.Index(value, " (in "); index >= 0 {
		frame.function = strings.TrimSpace(value[:index])
	} else {
		frame.function = value
	}
	for _, field := range strings.Fields(value) {
		if file, line := sourceLocation(field); line > 0 {
			frame.file, frame.line = file, line
		}
	}
	return frame
}

func temporaryProfile(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	return path, file.Close()
}
