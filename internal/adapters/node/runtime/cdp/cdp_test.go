package cdp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeResponse(t *testing.T) {
	var response struct {
		Value int `json:"value"`
	}
	require.NoError(t, decodeResponse("Runtime.evaluate", []byte(`{"id":1,"result":{"value":42}}`), &response))
	require.Equal(t, 42, response.Value)

	err := decodeResponse("Profiler.stop", []byte(`{"id":2,"error":{"message":"failed"}}`), nil)
	require.ErrorContains(t, err, "cdp Profiler.stop: failed")

	require.Error(t, decodeResponse("Runtime.evaluate", []byte(`{"id":`), &response))
}
