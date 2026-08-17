package runtime

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDecodeEventsReadsFixture decodes a real OCaml .events ring buffer.
func TestDecodeEventsReadsFixture(t *testing.T) {
	path := os.Getenv("SENBON_TEST_EVENTS")
	if path == "" {
		t.Skip("SENBON_TEST_EVENTS not set")
	}
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	counts, err := decodeEvents(data)
	require.NoError(t, err)
	t.Logf("minor=%d major=%d", counts.MinorCollections, counts.MajorCollections)
	require.NotZero(t, counts.MinorCollections+counts.MajorCollections)
}
