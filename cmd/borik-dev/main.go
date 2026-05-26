package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/bot"
	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/worker"
	"github.com/fogo-sh/borik/pkg/logging"
)

var temporalDevServer bool

var rootCmd = &cobra.Command{
	Use:   "borik-dev",
	Short: "Run the Borik Discord bot and image processing worker together",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		defer cancel()

		var temporalServer *testsuite.DevServer
		if temporalDevServer {
			log.Info().Msg("Starting embedded Temporal dev server")

			server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{
				ClientOptions: &client.Options{
					HostPort:  config.Instance.TemporalHostPort,
					Namespace: config.Instance.TemporalNamespace,
					Logger:    logging.NewTemporalLogger(),
				},
				EnableUI: true,
				LogLevel: "warn",
			})
			if err != nil {
				log.Fatal().Err(err).Msg("Error starting embedded Temporal dev server")
			}
			temporalServer = server
			config.Instance.TemporalHostPort = server.FrontendHostPort()

			log.Info().
				Str("host_port", config.Instance.TemporalHostPort).
				Str("namespace", config.Instance.TemporalNamespace).
				Msg("Embedded Temporal dev server is running")
		}

		imagick.Initialize()
		defer imagick.Terminate()

		w, err := worker.New()
		if err != nil {
			stopTemporalServer(temporalServer)
			log.Fatal().Err(err).Msg("Error creating worker")
		}

		borik, err := bot.New()
		if err != nil {
			stopTemporalServer(temporalServer)
			log.Fatal().Err(err).Msg("Error creating bot")
		}

		errCh := make(chan error, 2)
		go func() {
			if err := w.Start(); err != nil {
				errCh <- fmt.Errorf("worker stopped with error: %w", err)
			} else {
				errCh <- nil
			}
		}()
		go func() {
			if err := borik.Start(); err != nil {
				errCh <- fmt.Errorf("bot stopped with error: %w", err)
			} else {
				errCh <- nil
			}
		}()

		log.Info().Msg("Borik dev process is now running, press CTRL-C to exit.")

		var runErr error
		select {
		case <-ctx.Done():
			log.Info().Msg("Quitting Borik dev process")
		case err := <-errCh:
			runErr = err
			if err != nil {
				log.Error().Err(err).Msg("Borik dev process is stopping")
			}
		}

		stopComponent("bot", borik.Stop)
		stopComponent("worker", w.Stop)
		stopTemporalServer(temporalServer)

		if runErr != nil {
			log.Fatal().Err(runErr).Msg("Borik dev process exited with an error")
		}
	},
}

func init() {
	cobra.OnInitialize(loadConfig)
	rootCmd.Flags().BoolVar(
		&temporalDevServer,
		"temporal-dev-server",
		true,
		"start an embedded Temporal dev server before starting Borik",
	)
}

func loadConfig() {
	err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}
}

func stopComponent(name string, stop func()) {
	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Warn().Str("component", name).Msg("Timed out stopping component")
	}
}

func stopTemporalServer(server *testsuite.DevServer) {
	if server == nil {
		return
	}

	log.Info().Msg("Stopping embedded Temporal dev server")
	server.Client().Close()
	if err := server.Stop(); err != nil {
		log.Error().Err(err).Msg("Error stopping embedded Temporal dev server")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
