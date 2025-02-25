package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ip2dat",
	Short: "The IP .dat converter written in Go.",
	Long:  "The IP .dat converter written in Go.",
}

// Execute executes the root command.
func Execute() error { return rootCmd.Execute() }
