//go:build !darwin && !linux

package profiler

import (
	"context"
	"errors"

	"github.com/briheet/sen/internal/adapters/rust/analysis"
	"github.com/briheet/sen/internal/model"
)

func Capture(context.Context, int, *analysis.Symbols) (*model.Profile, *model.Trace, uint64, error) {
	return nil, nil, 0, errors.New("Rust profiling is supported only on macOS and Linux")
}
