package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/briheet/sen/internal/tui/styles"
	"github.com/spf13/cobra"
)

const (
	Version = "0.1.0"
	cpuFile = "cpu.pprof"
	memFile = "mem.pprof"

	ascii = `
    /$$$$$$                     
   /$$__  $$                    
  | $$  \__/  /$$$$$$  /$$$$$$$ 
  |  $$$$$$  /$$__  $$| $$__  $$
   \____  $$| $$$$$$$$| $$  \ $$
   /$$  \ $$| $$_____/| $$  | $$
  |  $$$$$$/|  $$$$$$$| $$  | $$
   \______/  \_______/|__/  |__/
`
)

var logoStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(styles.Zakura.Primary)

func Execute(ctx context.Context) int {
	var profile bool
	var cpuProfile *os.File
	rootCmd := &cobra.Command{
		Use:     "sen",
		Aliases: []string{"senbonzakura"},
		Short:   "Multi-service runtime analysis and visualization.",
		Long: strings.Trim(`
Inspired by Senbonzakura (千本桜), Sen is a multi service runtime
analysis and visualization tool. It combines source level program
analysis, runtime instrumentation, profiling, and service telemetry
to show how your application behaves in real time.

Model application processes and dependencies such as Redis and PostgreSQL
together as a single observable system.
`, "\n"),
		Example: `
# defaults to sen.toml
sen run ./config/

# pass in config file small flag
sen run -c ./config/sen.toml

# pass in config file long flag
sen run --config ./config/sen.toml
`,
		Args:    cobra.NoArgs,
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !profile {
				return nil
			}

			f, perr := os.Create(cpuFile)
			if perr != nil {
				return perr
			}
			if perr = pprof.StartCPUProfile(f); perr != nil {
				_ = f.Close()
				return perr
			}
			cpuProfile = f
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), logoStyle.Render(strings.Trim(ascii, "\n"))); err != nil {
				return err
			}
			return cmd.Help()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if !profile {
				return nil
			}

			pprof.StopCPUProfile()
			if err := cpuProfile.Close(); err != nil {
				return err
			}

			f, perr := os.Create(memFile)
			if perr != nil {
				return perr
			}
			defer func() { _ = f.Close() }()

			runtime.GC()
			err := pprof.WriteHeapProfile(f)
			return err
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&profile, "profile", "p", false, "record CPU and Mem pprof")
	rootCmd.AddCommand(RunCmd())
	if err := fang.Execute(ctx,
		rootCmd,
		fang.WithVersion(Version),
		fang.WithColorSchemeFunc(styles.FangColorScheme(styles.Zakura)),
	); err != nil {
		return 1
	}

	return 0
}
