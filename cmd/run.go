package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/fogo-sh/borik/pkg/bot"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Discord bot",
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

		log.Info().Msg("Borik is now running, press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
		log.Info().Msg("Quitting Borik")

		borik.Stop()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
