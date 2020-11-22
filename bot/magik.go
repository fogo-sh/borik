package bot

import (
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _MagikArgs struct {
	ImageURL string  `default:""`
	Scale    float64 `default:"1"`
}

func _MagikCommand(message *discordgo.MessageCreate, args _MagikArgs) {
	defer TypingIndicator(message)()

	if args.Scale == 0 {
		args.Scale = 1
	}

	if args.ImageURL == "" {
		var err error
		args.ImageURL, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	operationWrapper := func(srcBytes []byte, destBuffer io.Writer) error {
		return Magik(srcBytes, destBuffer, args.Scale)
	}
	PrepareAndInvokeOperation(message, args.ImageURL, operationWrapper)
}
