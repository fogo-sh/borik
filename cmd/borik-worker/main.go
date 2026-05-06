package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/worker"
)

var rootCmd = &cobra.Command{
	Use:   "borik-worker",
	Short: "Run the Borik image processing worker",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		imagick.Initialize()
		defer imagick.Terminate()

		workerCount, _ := cmd.Flags().GetUint("concurrency")

		var workers []*worker.Worker
		for range workerCount {
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
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		log.Info().Msg("Quitting Borik worker")

		for _, w := range workers {
			w.Stop()
		}
	},
}

func init() {
	cobra.OnInitialize(loadConfig)
	rootCmd.Flags().Uint("concurrency", 1, "Number of concurrent worker processes to run")
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
