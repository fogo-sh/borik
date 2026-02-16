package bot

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go/v3"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/config"
)

type AISessionMetadata struct {
	Seed      int
	SessionID string
	UserID    string
}

type AIParams interface {
	SetExtraFields(map[string]any)
}

func attachSessionMetadata(params AIParams, metadata AISessionMetadata) {
	params.SetExtraFields(map[string]any{
		"litellm_session_id": metadata.SessionID,
		"user":               metadata.UserID,
	})
}

type ImageGenArgs struct {
	Prompt string `description:"Prompt to generate an image for."`
}

func editImage(wand *imagick.MagickWand, args ImageEditArgs, metadata AISessionMetadata) (*imagick.MagickWand, error) {
	err := ShrinkMaintainAspectRatio(wand, 896, 896)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	imageBlob, err := wand.GetImageBlob()
	if err != nil {
		return nil, fmt.Errorf("error getting image blob: %w", err)
	}
	imageReader := bytes.NewReader(imageBlob)

	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, metadata.Seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	params := openai.ImageEditParams{
		Image: openai.ImageEditParamsImageUnion{
			OfFileArray: []io.Reader{imageReader},
		},
		Prompt: finalPrompt,
		Model:  config.Instance.OpenaiImageEditModel,
		// OpenAI's models require one of a few specific sizes, but stable-diffusion.cpp is more flexible
		// Pass the original image size to prevent it cropping it
		Size: openai.ImageEditParamsSize(fmt.Sprintf(
			"%dx%d",
			wand.GetImageWidth(),
			wand.GetImageHeight(),
		)),
		ResponseFormat: openai.ImageEditParamsResponseFormatB64JSON,
	}
	attachSessionMetadata(&params, metadata)

	editedImage, err := Instance.openAiClient.Images.Edit(
		context.TODO(),
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("error editing image: %w", err)
	}

	if len(editedImage.Data) == 0 {
		return nil, fmt.Errorf("no image data returned from edit")
	}

	decodedImg, err := base64.StdEncoding.DecodeString(editedImage.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("error decoding edited image: %w", err)
	}

	newWand := imagick.NewMagickWand()
	err = newWand.ReadImageBlob(decodedImg)
	if err != nil {
		return nil, fmt.Errorf("error reading edited image blob: %w", err)
	}

	return newWand, nil
}

type ImageEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
}

func (args ImageEditArgs) GetImageURL() string {
	return args.ImageURL
}

func ImageEdit(wand *imagick.MagickWand, args ImageEditArgs) ([]*imagick.MagickWand, error) {
	seed := rand.Int()
	editedImage, err := editImage(wand, args, AISessionMetadata{Seed: seed})
	if err != nil {
		return nil, err
	}

	return []*imagick.MagickWand{editedImage}, nil
}

type LoopEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
}

func (args LoopEditArgs) GetImageURL() string {
	return args.ImageURL
}

func LoopEdit(wand *imagick.MagickWand, args LoopEditArgs) ([]*imagick.MagickWand, error) {
	seed := rand.Int()
	metadata := AISessionMetadata{Seed: seed}
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps)

	currentWand := wand
	var err error
	for range args.Steps {
		metadata.Seed++ // Increment the seed for each iteration to produce different results
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt,
		}, metadata)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
}

func generateImage(ctx *OperationContext, args ImageGenArgs) {
	defer TypingIndicatorForContext(ctx)()

	if ctx.Interaction != nil {
		err := ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send deferred interaction response")
			return
		}
	}

	seed := rand.Int()
	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	params := openai.ImageGenerateParams{
		Prompt:         finalPrompt,
		Size:           "512x512",
		Model:          config.Instance.OpenaiImageGenModel,
		ResponseFormat: openai.ImageGenerateParamsResponseFormatB64JSON,
	}
	attachSessionMetadata(&params, AISessionMetadata{
		SessionID: ctx.GetSourceID(),
		UserID:    ctx.GetUserID(),
	})

	image, err := Instance.openAiClient.Images.Generate(
		context.TODO(),
		params,
	)
	if err != nil {
		ctx.RunCallbacks(
			func(m *discordgo.MessageCreate) {
				Instance.session.ChannelMessageSendReply(
					m.ChannelID,
					"Error generating image: `"+err.Error()+"`",
					m.Reference(),
				)
			},
			func(i *discordgo.InteractionCreate) {
				errMsg := "Error generating image: `" + err.Error() + "`"
				if _, editErr := ctx.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errMsg,
				}); editErr != nil {
					log.Error().Err(editErr).Msg("Failed to send error response")
				}
			},
		)
		return
	}

	file := &discordgo.File{
		Name:        "generated.png",
		ContentType: "image/png",
		Reader:      base64.NewDecoder(base64.StdEncoding, strings.NewReader(image.Data[0].B64JSON)),
	}

	ctx.RunCallbacks(
		func(m *discordgo.MessageCreate) {
			Instance.session.ChannelMessageSendComplex(
				m.ChannelID,
				&discordgo.MessageSend{
					Reference: m.Reference(),
					Files:     []*discordgo.File{file},
				},
			)
		},
		func(i *discordgo.InteractionCreate) {
			_, err := ctx.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Files: []*discordgo.File{file},
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to edit deferred interaction response")
			}
		},
	)
}

func ImageGenTextCommand(message *discordgo.MessageCreate, args ImageGenArgs) {
	generateImage(NewOperationContextFromMessage(Instance.session, message), args)
}

func ImageGenSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args ImageGenArgs) {
	generateImage(NewOperationContextFromInteraction(session, interaction), args)
}

type FlipFlopArgs struct {
	Prompt1  string `description:"First prompt to edit the image with."`
	Prompt2  string `description:"Second prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
}

func (args FlipFlopArgs) GetImageURL() string {
	return args.ImageURL
}

func FlipFlop(wand *imagick.MagickWand, args FlipFlopArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps*2+1)

	editedFrames = append(editedFrames, wand)

	currentWand := wand
	var err error
	for range args.Steps {
		metadata.Seed++ // Increment the seed for each iteration to produce different results
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt1,
		}, metadata)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt2,
		}, metadata)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
}

func FlipFlopCommand(message *discordgo.MessageCreate, args FlipFlopArgs) {
	seed := rand.Int()
	PrepareAndInvokeOperation(message, args, func(wand *imagick.MagickWand, args FlipFlopArgs) ([]*imagick.MagickWand, error) {
		return FlipFlop(wand, args, AISessionMetadata{
			Seed:      seed,
			SessionID: message.ID,
			UserID:    message.Author.ID,
		})
	})
}
