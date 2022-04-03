package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _MaltArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degree   float64 `default:"45" description:"Number of degrees to rotate the image by while processing."`
}

func _MaltCommand(message *discordgo.MessageCreate, args _MaltArgs) {
	defer TypingIndicator(message)()

	if args.ImageURL == "" {
		var err error
		args.ImageURL, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	PrepareAndInvokeOperation(message, args.ImageURL, args, Malt)
}
