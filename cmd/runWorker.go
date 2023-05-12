package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/fogo-sh/borik/pkg/jobs/worker"
)

var runWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run the Borik worker",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		w, err := worker.New()
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating bot")
		}

		go func() {
			err := w.Start()
			if err != nil {
				log.Fatal().Err(err).Msg("Error starting worker")
			}
		}()

		log.Info().Msg("Borik worker is now running, press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
		log.Info().Msg("Quitting Borik worker")

		w.Stop()
	},
}

func init() {
	runCmd.AddCommand(runWorkerCmd)
}
