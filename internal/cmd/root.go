package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/cobra"
)

const (
	Version = "0.1.0"
	CPUFile = "cpu.pprof"
	MEMFile = "mem.pprof"
	Ascii   = `
  /$$$$$$                      /$$                          
 /$$__  $$                    | $$                          
| $$  \__/  /$$$$$$  /$$$$$$$ | $$$$$$$   /$$$$$$  /$$$$$$$ 
|  $$$$$$  /$$__  $$| $$__  $$| $$__  $$ /$$__  $$| $$__  $$
 \____  $$| $$$$$$$$| $$  \ $$| $$  \ $$| $$  \ $$| $$  \ $$
 /$$  \ $$| $$_____/| $$  | $$| $$  | $$| $$  | $$| $$  | $$
|  $$$$$$/|  $$$$$$$| $$  | $$| $$$$$$$/|  $$$$$$/| $$  | $$
 \______/  \_______/|__/  |__/|_______/  \______/ |__/  |__/
`
)

var (
	Profile bool
)

func Execute(ctx context.Context) int {
	rootCmd := &cobra.Command{
		Use:     "senbon",
		Aliases: []string{"sen"},
		Short:   "Dynamic Program Analysis & Runtime Visualization of Golang programs.",
		Long: `
Dynamic Control-Flow Analysis and Runtime Visualization for Go Programs,
combining source-level program analysis, runtime instrumentation, profiling,
and execution-path visualization to show how code executes in real time.
`,
		Example: `
senbon ./<main.go-dir>
`,
		Args:    cobra.NoArgs,
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !Profile {
				return nil
			}

			f, perr := os.Create(CPUFile)
			if perr != nil {
				return perr
			}

			_ = pprof.StartCPUProfile(f)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(Ascii)
			return cmd.Help()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if !Profile {
				return nil
			}

			pprof.StopCPUProfile()

			f, perr := os.Create("mem.pprof")
			if perr != nil {
				return perr
			}
			defer f.Close()

			runtime.GC()
			err := pprof.WriteHeapProfile(f)
			return err
		},
	}

	// Define all your flags here
	rootCmd.PersistentFlags().BoolVarP(&Profile, "profile", "p", false, "record CPU and Mem pprof")

	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
