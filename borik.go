package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/saturn-sh/borik/bot"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("Failed to load .env file")
	}

	imagick.Initialize()
	defer imagick.Terminate()

	borik, err := bot.New()
	if err != nil {
		fmt.Printf("Error creating Borik instance: %s\n", err.Error())
		return
	}

	log.Debug().Msg("Opening Discord connection")
	err = borik.Session.Open()
	if err != nil {
		log.Error().Err(err).Msg("Error opening connection")
		return
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
}
