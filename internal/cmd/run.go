package cmd

import (
	"github.com/briheet/senbon/internal/engine"
	"github.com/spf13/cobra"
)

func RunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <path>",
		Short: "Build and run the application",
		Long: `Build and run a Go application under Senbon.

Senbon loads the target program, performs static analysis,
builds the program, starts runtime instrumentation, and
launches the TUI.`,
		Example: "senbon run ./cmd/server",
		Args:    cobra.ExactArgs(1),
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get sourcepath, init a new engine, run it asf.
			sourcePath := args[0]
			engine, err := engine.NewEngine(cmd.Context(), sourcePath)
			if err != nil {
				return err
			}
			defer engine.Runtime.Process.Cleanup()
			return engine.Run()
		},
	}

	return runCmd
}
