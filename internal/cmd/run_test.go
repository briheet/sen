package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmdConfigFlag(t *testing.T) {
	command := RunCmd()
	flag := command.Flags().Lookup("config")
	require.NotNil(t, flag)
	require.Equal(t, "c", flag.Shorthand)
	require.Equal(t, defaultConfigPath, flag.DefValue)
	require.NoError(t, command.Args(command, nil))
	require.NoError(t, command.Args(command, []string{"./project"}))
	require.Error(t, command.Args(command, []string{"one", "two"}))
	require.NoError(t, command.ParseFlags([]string{"-c", "custom.toml"}))
	require.Equal(t, "custom.toml", flag.Value.String())
}
