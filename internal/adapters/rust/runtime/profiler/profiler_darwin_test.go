//go:build darwin

package profiler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSamplePreservesTreeDepth(t *testing.T) {
	t.Parallel()
	input := `Call graph:
    92 Thread_1
    + 92 root
    +   60 branch_a
    +     60 leaf_a
    +   32 branch_b
    +     32 leaf_b
    92 Thread_2
    + 92 second_root
    +   92 second_leaf
Total number in stack
`

	samples, err := parseSample(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, samples, 3)
	require.Equal(t, []uint64{60, 32, 92}, []uint64{samples[0].weight, samples[1].weight, samples[2].weight})
	require.Equal(t, "leaf_a", samples[0].frames[0].function)
	require.Equal(t, "root", samples[0].frames[2].function)
}
