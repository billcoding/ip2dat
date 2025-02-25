package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v", "ver"},
	Short:   "Print the version",
	Long: `Print the version of ip2dat.
Simply type fly help version for full details.`,
	Example: `ip2dat v|ver|version [-a]`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVersion(goVersionFlag)
	},
}

func init() {
	versionCmd.PersistentFlags().BoolVarP(&goVersionFlag, "all", "a", false, "print Go SDK version")
	rootCmd.AddCommand(versionCmd)
}

var goVersionFlag = false

const version = "v1.0.0"

func PrintVersion(goVersion bool) {
	_, _ = fmt.Fprintln(os.Stdout, version)
	if goVersion {
		_, _ = fmt.Fprintln(os.Stdout, runtime.Version())
	}
}
