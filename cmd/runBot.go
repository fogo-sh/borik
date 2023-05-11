package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/bot"
)

var runBotCmd = &cobra.Command{
	Use:   "bot",
	Short: "Run the Discord bot",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		imagick.Initialize()
		defer imagick.Terminate()

		borik, err := bot.New()
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating bot")
		}

		defer func() {
			err := borik.Trace.Shutdown(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("Error shutting down trace provider")
			}
		}()

		log.Debug().Msg("Opening Discord connection")
		err = borik.Session.Open()
		if err != nil {
			log.Fatal().Err(err).Msg("Error opening Discord connection")
		}

		log.Info().Msg("Borik is now running, press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
		log.Info().Msg("Quitting Borik")

		err = borik.Session.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing Discord connection")
		}
	},
}

func init() {
	runCmd.AddCommand(runBotCmd)
}
