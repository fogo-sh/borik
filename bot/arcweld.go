package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _ArcweldArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
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

	PrepareAndInvokeOperation(message, args.ImageURL, args, Arcweld)
}
