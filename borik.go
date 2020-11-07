package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/saturn-sh/borik/bot"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	borik, err := bot.New()
	if err != nil {
		fmt.Printf("Error creating Borik instance: %s\n", err.Error())
		return
	}

	borik.Session.AddHandler(messageCreate)

	borik.Session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	config := bot.Instance.Config
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, config.Prefix) {
		return
	}

	imageURI, err := bot.ImageURIFromCommand(s, m, config.Prefix)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	fmt.Println("found imageURI to borik: ", imageURI)

	srcBytes, err := bot.DownloadImage(imageURI)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}
	destBuffer := new(bytes.Buffer)

	log.Debug().Msg("Beginning processing image")
	err = bot.Magik(srcBytes, destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process image")
		return
	}

	log.Debug().Msg("Image processed, uploading result")
	_, err = s.ChannelFileSend(m.ChannelID, "test.jpeg", destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
	}
}
