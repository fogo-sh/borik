package bot

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/fogo-sh/borik/pkg/jobs/args"
)

func MakeWorkflowTextCommand[K args.JobArgs]() func(*discordgo.MessageCreate, K) {
	return func(message *discordgo.MessageCreate, args K) {
		PrepareAndInvokeWorkflow(NewOperationContextFromMessage(Instance.session, message), args)
	}
}

func MakeWorkflowSlashCommand[K args.JobArgs]() func(*discordgo.Session, *discordgo.InteractionCreate, K) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, args K) {
		PrepareAndInvokeWorkflow(NewOperationContextFromInteraction(session, interaction), args)
	}
}

func PrepareAndInvokeWorkflow[K args.JobArgs](ctx *OperationContext, cmdArgs K) {
	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	imageURL := cmdArgs.GetImageURL()
	if imageURL == "" {
		var err error
		imageURL, err = ctx.FindImageURL()
		if err != nil {
			if sendErr := ctx.SendText("Error finding image URL: " + err.Error()); sendErr != nil {
				log.Error().Err(sendErr).Msg("Failed to send error message")
			}
			return
		}
	}

	target := ctx.DeliveryTarget("Error processing image")
	target.FilenameBase = strings.ToLower(cmdArgs.ActivityName())

	err := Instance.triggerJob(context.Background(), ctx.GetSourceID(), imageURL, cmdArgs, target)
	if err != nil {
		if sendErr := ctx.SendText("Error triggering jobs: " + err.Error()); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
		return
	}
}
