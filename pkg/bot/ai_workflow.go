package bot

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/fogo-sh/borik/pkg/jobs/args"
)

type ImageGenArgs struct {
	Prompt string `description:"Prompt to generate an image for."`
}

type ImageEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
}

type LoopEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
}

type FlipFlopArgs struct {
	Prompt1  string `description:"First prompt to edit the image with."`
	Prompt2  string `description:"Second prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
}

type AiZoomArgs struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
	Steps    uint   `default:"2" description:"Number of zoom steps to perform."`
}

type AiLoopZoomArgs struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
	Steps    uint   `default:"5" description:"Number of zoom steps to perform."`
}

func buildAIMetadata(ctx *OperationContext) args.AIMetadata {
	return args.AIMetadata{
		Seed:      rand.Int(),
		SessionID: ctx.GetSourceID(),
		UserID:    ctx.GetUserID(),
	}
}

func ImageGenWorkflowTextCommand(message *discordgo.MessageCreate, imageGenArgs ImageGenArgs) {
	ctx := NewOperationContextFromMessage(Instance.session, message)
	PrepareAndInvokeGenerateImage(ctx, args.ImageGen{
		Prompt:   imageGenArgs.Prompt,
		Metadata: buildAIMetadata(ctx),
	})
}

func ImageGenWorkflowSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, imageGenArgs ImageGenArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)
	PrepareAndInvokeGenerateImage(ctx, args.ImageGen{
		Prompt:   imageGenArgs.Prompt,
		Metadata: buildAIMetadata(ctx),
	})
}

func ImageEditWorkflowTextCommand(message *discordgo.MessageCreate, imageEditArgs ImageEditArgs) {
	ctx := NewOperationContextFromMessage(Instance.session, message)
	PrepareAndInvokeWorkflow(ctx, args.ImageEdit{
		Prompt:   imageEditArgs.Prompt,
		ImageURL: imageEditArgs.ImageURL,
		Metadata: buildAIMetadata(ctx),
	})
}

func ImageEditWorkflowSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, imageEditArgs ImageEditArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)
	PrepareAndInvokeWorkflow(ctx, args.ImageEdit{
		Prompt:   imageEditArgs.Prompt,
		ImageURL: imageEditArgs.ImageURL,
		Metadata: buildAIMetadata(ctx),
	})
}

func LoopEditWorkflowTextCommand(message *discordgo.MessageCreate, loopEditArgs LoopEditArgs) {
	ctx := NewOperationContextFromMessage(Instance.session, message)
	PrepareAndInvokeWorkflow(ctx, args.LoopEdit{
		Prompt:   loopEditArgs.Prompt,
		ImageURL: loopEditArgs.ImageURL,
		Steps:    loopEditArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func LoopEditWorkflowSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, loopEditArgs LoopEditArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)
	PrepareAndInvokeWorkflow(ctx, args.LoopEdit{
		Prompt:   loopEditArgs.Prompt,
		ImageURL: loopEditArgs.ImageURL,
		Steps:    loopEditArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func FlipFlopWorkflowTextCommand(message *discordgo.MessageCreate, flipFlopArgs FlipFlopArgs) {
	ctx := NewOperationContextFromMessage(Instance.session, message)
	PrepareAndInvokeWorkflow(ctx, args.FlipFlop{
		Prompt1:  flipFlopArgs.Prompt1,
		Prompt2:  flipFlopArgs.Prompt2,
		ImageURL: flipFlopArgs.ImageURL,
		Steps:    flipFlopArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func FlipFlopWorkflowSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, flipFlopArgs FlipFlopArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)
	PrepareAndInvokeWorkflow(ctx, args.FlipFlop{
		Prompt1:  flipFlopArgs.Prompt1,
		Prompt2:  flipFlopArgs.Prompt2,
		ImageURL: flipFlopArgs.ImageURL,
		Steps:    flipFlopArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func AiZoomWorkflowTextCommand(message *discordgo.MessageCreate, aiZoomArgs AiZoomArgs) {
	ctx := NewOperationContextFromMessage(Instance.session, message)
	PrepareAndInvokeWorkflow(ctx, args.AiZoom{
		ImageURL: aiZoomArgs.ImageURL,
		Prompt:   aiZoomArgs.Prompt,
		Steps:    aiZoomArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func AiZoomWorkflowSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, aiZoomArgs AiZoomArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)
	PrepareAndInvokeWorkflow(ctx, args.AiZoom{
		ImageURL: aiZoomArgs.ImageURL,
		Prompt:   aiZoomArgs.Prompt,
		Steps:    aiZoomArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func AiLoopZoomWorkflowTextCommand(message *discordgo.MessageCreate, aiLoopZoomArgs AiLoopZoomArgs) {
	ctx := NewOperationContextFromMessage(Instance.session, message)
	PrepareAndInvokeWorkflow(ctx, args.AiLoopZoom{
		ImageURL: aiLoopZoomArgs.ImageURL,
		Prompt:   aiLoopZoomArgs.Prompt,
		Steps:    aiLoopZoomArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func AiLoopZoomWorkflowSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, aiLoopZoomArgs AiLoopZoomArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)
	PrepareAndInvokeWorkflow(ctx, args.AiLoopZoom{
		ImageURL: aiLoopZoomArgs.ImageURL,
		Prompt:   aiLoopZoomArgs.Prompt,
		Steps:    aiLoopZoomArgs.Steps,
		Metadata: buildAIMetadata(ctx),
	})
}

func PrepareAndInvokeGenerateImage(ctx *OperationContext, imageGenArgs args.ImageGen) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	resultFormat, resultReader, err := Instance.triggerGenerateImage(context.Background(), ctx.GetSourceID(), imageGenArgs)
	if err != nil {
		if sendErr := ctx.SendText("Error triggering jobs: " + err.Error()); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
		return
	}

	err = ctx.SendFiles([]*discordgo.File{
		{
			Name:   fmt.Sprintf("generated.%s", strings.ToLower(resultFormat)),
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
