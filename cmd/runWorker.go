package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/worker"
)

var runWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run the Borik worker",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		imagick.Initialize()
		defer imagick.Terminate()

		workerCount, _ := cmd.Flags().GetUint("concurrency")

		var workers []*worker.Worker

		for i := uint(0); i < workerCount; i++ {
			w, err := worker.New()
			if err != nil {
				log.Fatal().Err(err).Msg("Error creating worker")
			}

			workers = append(workers, w)

			go func() {
				err := w.Start()
				if err != nil {
					log.Fatal().Err(err).Msg("Error starting worker")
				}
			}()
		}

		log.Info().Msg("Borik worker is now running, press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
		log.Info().Msg("Quitting Borik worker")

		for _, w := range workers {
			w.Stop()
		}
	},
}

func init() {
	runWorkerCmd.Flags().Uint("concurrency", 1, "Number of concurrent worker processes to run")
	runCmd.AddCommand(runWorkerCmd)
}
