package bot

import (
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type _DeepfryArgs struct {
	ImageURL        string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius      float64 `default:"100" description:"Radius of outline to draw around edges."`
	DownscaleFactor uint    `default:"2" description:"Factor to downscale the image by while processing."`
}

func _DeepfryCommand(message *discordgo.MessageCreate, args _DeepfryArgs) {
	if args.ImageURL == "pipeline" {
		err := Instance.PipelineManager.AddStep(message, "deepfry", args)
		if err != nil {
			Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror adding step to pipeline: %s\n```", err.Error()))
		}
		Instance.Session.ChannelMessageSend(message.ChannelID, "Step added to pipeline.")
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
		return Deepfry(srcBytes, destBuffer, args)
	}
	PrepareAndInvokeOperation(message, args.ImageURL, operationWrapper)
}
