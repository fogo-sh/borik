package bot

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.temporal.io/sdk/client"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/workflows"
)

type testArgs struct {
	ImageURL         string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale            float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
	WidthMultiplier  float64 `default:"0.5" description:"Multiplier to apply to the width of the input image to produce the intermediary image."`
	HeightMultiplier float64 `default:"0.5" description:"Multiplier to apply to the height of the input image to produce the intermediary image."`
}

func (b *Bot) workflowTestCommand(message *discordgo.MessageCreate, args testArgs) {
	if args.ImageURL == "" {
		imageUrl, err := FindImageURL(message)
		if err != nil {
			b.session.ChannelMessageSend(message.ChannelID, "Error finding image URL: "+err.Error())
			return
		}
		args.ImageURL = imageUrl
	}

	we, err := b.temporalClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        "test-workflow",
			TaskQueue: config.Instance.TemporalQueueName,
		},
		workflows.ProcessImageWorkflow,
		workflows.ProcessImageArgs{
			ImageURL:     args.ImageURL,
			ActivityName: "Magik",
			ActivityArgs: activities.MagikArgs{
				Scale:            args.Scale,
				WidthMultiplier:  args.WidthMultiplier,
				HeightMultiplier: args.HeightMultiplier,
			},
		},
	)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error executing workflow: "+err.Error())
		return
	}

	var result workflows.ProcessedImageResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error getting workflow result: "+err.Error())
		return
	}

	image, err := result.Workspace.Retrieve(result.Image)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error retrieving image: "+err.Error())
		return
	}

	resultBuf := bytes.NewBuffer(image)

	// Send result as a file attachment
	_, err = b.session.ChannelFileSend(message.ChannelID, fmt.Sprintf("output.%s", result.Format), resultBuf)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error sending result: "+err.Error())
		return
	}
}
