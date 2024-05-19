package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Borik's components",
}

func init() {
	rootCmd.AddCommand(runCmd)
}
