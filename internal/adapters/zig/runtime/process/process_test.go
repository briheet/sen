package process

import (
	"testing"

	"github.com/briheet/senbon/internal/adapters/zig/analysis"
	"github.com/stretchr/testify/require"
)

func TestBuildArgs(t *testing.T) {
	project := &analysis.Project{
		Entry: "/proj/src/main.zig",
		Modules: map[string]string{
			"router":   "/proj/src/router.zig",
			"services": "/proj/src/services.zig",
		},
		Imports: map[string][]string{
			"/proj/src/main.zig":     {"router", "services"},
			"/proj/src/router.zig":   {"services"},
			"/proj/src/services.zig": nil,
		},
	}

	args := buildArgs(project, "/proj", "/tmp/sampler.zig", "/tmp/main", "/tmp/cache")
	require.Contains(t, args, "-lc")
	require.Contains(t, args, "--dep")
	require.Contains(t, args, "user=user")
	require.Contains(t, args, "-Mroot=/tmp/sampler.zig")
	require.Contains(t, args, "-Muser=/proj/src/main.zig")
	require.Contains(t, args, "-Mrouter=/proj/src/router.zig")
	require.Contains(t, args, "-Mservices=/proj/src/services.zig")
	require.Contains(t, args, "-femit-bin=/tmp/main")
	require.Contains(t, args, "-O")
	require.Contains(t, args, "Debug")

	// entry module receives its imports before its -M
	userIndex := indexOf(args, "-Muser=/proj/src/main.zig")
	require.Greater(t, userIndex, 0)
	require.Equal(t, "--dep", args[userIndex-4])
	require.Equal(t, "router=router", args[userIndex-3])
	require.Equal(t, "--dep", args[userIndex-2])
	require.Equal(t, "services=services", args[userIndex-1])
}

func indexOf(slice []string, value string) int {
	for index, item := range slice {
		if item == value {
			return index
		}
	}
	return -1
}
