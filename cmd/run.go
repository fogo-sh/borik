package cmd

import "github.com/spf13/cobra"

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run components of Borik",
}

func init() {
	rootCmd.AddCommand(runCmd)
}
