package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/fogo-sh/borik/pkg/bot"
	"github.com/fogo-sh/borik/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "borik",
	Short: "Run the Borik Discord bot",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		borik, err := bot.New()
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating bot")
		}

		go func() {
			err := borik.Start()
			if err != nil {
				log.Fatal().Err(err).Msg("Error starting bot")
			}
		}()

		log.Info().Msg("Borik bot is now running, press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		log.Info().Msg("Quitting Borik bot")

		borik.Stop()
	},
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
