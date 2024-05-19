package cmd

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/fogo-sh/borik/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "borik",
	Short: "Discord bot for destroying images",
}

func init() {
	cobra.OnInitialize(loadConfig)
}

func loadConfig() {
	err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
