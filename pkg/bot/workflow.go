package bot

import (
	"context"
	"fmt"
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
	defer TypingIndicatorForContext(ctx)()

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

	resultFormat, resultReader, err := Instance.triggerJob(context.Background(), ctx.GetSourceID(), imageURL, cmdArgs)
	if err != nil {
		if sendErr := ctx.SendText("Error triggering jobs: " + err.Error()); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
		return
	}

	err = ctx.SendFiles([]*discordgo.File{
		{
			Name:   fmt.Sprintf("%s.%s", strings.ToLower(cmdArgs.ActivityName()), resultFormat),
			Reader: resultReader,
		},
	})
	if err != nil {
		if sendErr := ctx.SendText("Error sending result: " + err.Error()); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
		return
	}
}
