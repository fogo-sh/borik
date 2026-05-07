package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/fogo-sh/borik/pkg/jobs/args"
)

func (b *Bot) workflowTestCommand(message *discordgo.MessageCreate, cmdArgs args.Magik) {
	if cmdArgs.ImageURL == "" {
		imageUrl, err := FindImageURLFromMessage(message)
		if err != nil {
			b.session.ChannelMessageSend(message.ChannelID, "Error finding image URL: "+err.Error())
			return
		}
		cmdArgs.ImageURL = imageUrl
	}

	resultFormat, resultReader, err := b.triggerJobs(context.Background(), cmdArgs)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error triggering jobs: "+err.Error())
		return
	}

	// Send result as a file attachment
	_, err = b.session.ChannelFileSend(message.ChannelID, fmt.Sprintf("output.%s", resultFormat), resultReader)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error sending result: "+err.Error())
		return
	}
}
