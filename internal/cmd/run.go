package cmd

import (
	"github.com/briheet/senbon/internal/engine"
	"github.com/spf13/cobra"
)

func RunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <language> <path>",
		Short: "Analyze and run an application",
		Long: `Build and run an application.

Senbon loads the target program, performs static analysis,
builds the program, starts runtime instrumentation, and
launches the TUI.`,
		Example: "senbon run node ./examples/node",
		Args:    cobra.ExactArgs(2),
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := engine.NewEngine(cmd.Context(), args[1], args[0])
			if err != nil {
				return err
			}
			defer func() { _ = target.Cleanup() }()
			return target.Run()
		},
	}
	return runCmd
}
