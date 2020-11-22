package bot

import (
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _ArcweldArgs struct {
	ImageURL string `default:""`
}

func _ArcweldCommand(message *discordgo.MessageCreate, args _ArcweldArgs) {
	defer TypingIndicator(message)()

	if args.ImageURL == "" {
		var err error
		args.ImageURL, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	operationWrapper := func(srcBytes []byte, destBuffer io.Writer) error {
		return Arcweld(srcBytes, destBuffer)
	}
	PrepareAndInvokeOperation(message, args.ImageURL, operationWrapper)
}
