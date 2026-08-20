package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLine(t *testing.T) {
	t.Parallel()
	sample, err := ParseLine("tb.replica_request_us.count:42|c|#cluster:0123456789abcdef0123456789abcdef,replica:1,operation:create_accounts")
	require.NoError(t, err)
	require.Equal(t, "replica_request_us.count", sample.Name)
	require.Equal(t, float64(42), sample.Value)
	require.Equal(t, "c", sample.Type)
	require.Equal(t, uint32(1), sample.Replica)
	require.Equal(t, "create_accounts", sample.Operation)
}

func TestParseLineRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	for _, line := range []string{
		"other.metric:1|g|#cluster:0123456789abcdef0123456789abcdef,replica:0",
		"tb.replica_status:1|ms|#cluster:0123456789abcdef0123456789abcdef,replica:0",
		"tb.replica_status:1|g|#replica:0",
		"tb.replica_status:-1|g|#cluster:0123456789abcdef0123456789abcdef,replica:0",
	} {
		_, err := ParseLine(line)
		require.Error(t, err, line)
	}
}
