package trace

import (
	"bytes"
	"context"
	"maps"
	"runtime/trace"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadRuntimeTrace(t *testing.T) {
	data := runtimeTrace(t, "example.com/app")

	result, err := Read(bytes.NewReader(data))
	require.NoError(t, err)
	require.Empty(t, result.Events)
	require.NotEmpty(t, result.Stacks)
	require.Positive(t, result.Duration)
	require.NotNil(t, result.Aggregate)
	require.Equal(t, result.Duration, result.Aggregate.Summary.Duration)
	require.NotEmpty(t, result.Aggregate.Stacks)
	require.Positive(t, result.Aggregate.Summary.Goroutines)
}

func TestDecoderReusesAlternatingBuffers(t *testing.T) {
	firstData := runtimeTrace(t, "first")
	secondData := runtimeTrace(t, "second")
	thirdData := runtimeTrace(t, "third")
	var decoder Decoder

	first, err := decoder.Read(bytes.NewReader(firstData))
	require.NoError(t, err)
	firstAggregate := first.Aggregate
	firstStacks := maps.Clone(first.Aggregate.Stacks)
	firstDuration := first.Duration

	second, err := decoder.Read(bytes.NewReader(secondData))
	require.NoError(t, err)
	require.NotSame(t, first, second)
	require.NotSame(t, first.Aggregate, second.Aggregate)
	require.Equal(t, firstStacks, first.Aggregate.Stacks, "decoding must not mutate the active buffer")
	require.Equal(t, firstDuration, first.Duration)

	third, err := decoder.Read(bytes.NewReader(thirdData))
	require.NoError(t, err)
	require.Same(t, first, third)
	require.Same(t, firstAggregate, third.Aggregate)
	require.NotEmpty(t, third.Aggregate.Stacks)
}

func TestDecoderKeepsActiveTraceOnError(t *testing.T) {
	var decoder Decoder
	current, err := decoder.Read(bytes.NewReader(runtimeTrace(t, "current")))
	require.NoError(t, err)
	stacks := maps.Clone(current.Aggregate.Stacks)
	duration := current.Duration

	_, err = decoder.Read(bytes.NewReader([]byte("invalid trace")))
	require.Error(t, err)
	require.Equal(t, stacks, current.Aggregate.Stacks)
	require.Equal(t, duration, current.Duration)
}

func BenchmarkDecoder(b *testing.B) {
	data := runtimeTrace(b, "benchmark")
	b.Run("stateless", func(b *testing.B) {
		for range b.N {
			if _, err := Read(bytes.NewReader(data)); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("reused", func(b *testing.B) {
		var decoder Decoder
		for range b.N {
			if _, err := decoder.Read(bytes.NewReader(data)); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func runtimeTrace(tb testing.TB, message string) []byte {
	tb.Helper()
	var data bytes.Buffer
	require.NoError(tb, trace.Start(&data))

	ctx, task := trace.NewTask(context.Background(), "build")
	trace.WithRegion(ctx, "compile", func() {
		trace.Log(ctx, "package", message)
	})
	task.End()
	trace.Stop()
	return data.Bytes()
}
