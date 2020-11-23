package bot

import (
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _MaltArgs struct {
	ImageURL string  `default:""`
	Degree   float64 `default:"45"`
}

func _MaltCommand(message *discordgo.MessageCreate, args _MaltArgs) {
	if args.ImageURL == "pipeline" {
		err := Instance.PipelineManager.AddStep(message, "malt", args)
		if err != nil {
			Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror adding step to pipeline: %s\n```", err.Error()))
		}
		return
	}

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
		return Malt(srcBytes, destBuffer, args)
	}
	PrepareAndInvokeOperation(message, args.ImageURL, operationWrapper)
}
