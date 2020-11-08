package bot

import (
	"bytes"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _ArcweldArgs struct {
	ImageURL string `default:""`
}

func _ArcweldCommand(message *discordgo.MessageCreate, args _ArcweldArgs) {
	if args.ImageURL == "" {
		var err error
		args.ImageURL, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	srcBytes, err := DownloadImage(args.ImageURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}
	destBuffer := new(bytes.Buffer)

	log.Debug().Msg("Beginning processing image")
	err = Arcweld(srcBytes, destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process image")
		return
	}

	log.Debug().Msg("Image processed, uploading result")
	_, err = Instance.Session.ChannelFileSend(message.ChannelID, "test.jpeg", destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		_, err = Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
	}
}
