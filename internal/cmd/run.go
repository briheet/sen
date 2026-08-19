package cmd

import (
	"github.com/briheet/sen/internal/config"
	"github.com/briheet/sen/internal/tui"
	"github.com/spf13/cobra"
)

const defaultConfigPath = config.DefaultPath

// RunCmd loads configured services and runs their supported engines.
func RunCmd() *cobra.Command {
	var configPath string
	runCmd := &cobra.Command{
		Use:   "run [path]",
		Short: "Analyze and run a multi-service project",
		Long: `Load, analyze, and run services defined in a sen project.

Application runtimes and supporting services are loaded from configuration
and brought together for unified analysis and visualization.`,
		Example: "sen run\n  sen run ./examples/go/http\n  sen run --config ./config/sen.toml\n  sen run -c ./config",
		Args:    cobra.MaximumNArgs(1),
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.ResolvePath(configPath, cmd.Flags().Changed("config"), args)
			if err != nil {
				return err
			}
			configuration, err := config.Load(path)
			if err != nil {
				return err
			}
			application, err := tui.NewTui(cmd.Context(), configuration)
			if err != nil {
				return err
			}
			return application.Run(cmd.Context())
		},
	}
	runCmd.Flags().StringVarP(&configPath, "config", "c", defaultConfigPath, "path to sen.toml")
	return runCmd
}
