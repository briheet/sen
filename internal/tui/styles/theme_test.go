package styles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		name string
		want Theme
	}{
		{name: "", want: Zakura},
		{name: "zakura", want: Zakura},
		{name: "catppuccin-mocha", want: CatppuccinMocha},
		{name: "miyazaki-16", want: Miyazaki16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveTheme(test.name)
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}

	_, err := ResolveTheme("missing")
	require.ErrorContains(t, err, `unknown theme "missing"`)
}
