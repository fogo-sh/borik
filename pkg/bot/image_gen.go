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

func generateImage(ctx *OperationContext, args ImageGenArgs) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
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
		if sendErr := ctx.SendText("Error generating image: `" + err.Error() + "`"); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error response")
		}
		return
	}

	file := &discordgo.File{
		Name:        "generated.png",
		ContentType: "image/png",
		Reader:      base64.NewDecoder(base64.StdEncoding, strings.NewReader(image.Data[0].B64JSON)),
	}

	if err := ctx.SendFiles([]*discordgo.File{file}); err != nil {
		log.Error().Err(err).Msg("Failed to send generated image")
	}
}

func ImageGenTextCommand(message *discordgo.MessageCreate, args ImageGenArgs) {
	generateImage(NewOperationContextFromMessage(Instance.session, message), args)
}

func ImageGenSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args ImageGenArgs) {
	generateImage(NewOperationContextFromInteraction(session, interaction), args)
}

func editImage(wand *imagick.MagickWand, args ImageEditArgs, metadata AISessionMetadata, mask *imagick.MagickWand) (*imagick.MagickWand, error) {
	err := ShrinkMaintainAspectRatio(wand, 896, 896)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	imageBlob, err := wand.GetImageBlob()
	if err != nil {
		return nil, fmt.Errorf("error getting image blob: %w", err)
	}
	imageReader := bytes.NewReader(imageBlob)

	var maskReader io.Reader
	if mask != nil {
		err = ShrinkMaintainAspectRatio(mask, 896, 896)
		if err != nil {
			return nil, fmt.Errorf("error resizing mask: %w", err)
		}

		maskBlob, err := mask.GetImageBlob()
		if err != nil {
			return nil, fmt.Errorf("error getting mask blob: %w", err)
		}
		maskReader = bytes.NewReader(maskBlob)
	}

	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, metadata.Seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	params := openai.ImageEditParams{
		Image: openai.ImageEditParamsImageUnion{
			OfFileArray: []io.Reader{imageReader},
		},
		Mask:   maskReader,
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

func ImageEdit(wand *imagick.MagickWand, args ImageEditArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	editedImage, err := editImage(wand, args, metadata, nil)
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

func LoopEdit(wand *imagick.MagickWand, args LoopEditArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps)

	currentWand := wand
	var err error
	for range args.Steps {
		metadata.Seed++ // Increment the seed for each iteration to produce different results
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt,
		}, metadata, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
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
		}, metadata, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt2,
		}, metadata, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
}

type AiZoomArgs struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
}

func (args AiZoomArgs) GetImageURL() string {
	return args.ImageURL
}

func AiZoom(wand *imagick.MagickWand, args AiZoomArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	originalWidth := wand.GetImageWidth()
	originalHeight := wand.GetImageHeight()

	// Resize the image to be smaller, then let the model edit it back to the original size, which should have a zoom effect
	err := ShrinkMaintainAspectRatio(wand, uint(float64(wand.GetImageWidth())*0.8), uint(float64(wand.GetImageHeight())*0.8))
	if err != nil {
		return nil, fmt.Errorf("error resizing image for zoom: %w", err)
	}

	// Place the shrunk image in the center of a transparent canvas of the original size
	canvas := imagick.NewMagickWand()

	err = canvas.SetBackgroundColor(imagick.NewPixelWand()) // Transparent background
	if err != nil {
		return nil, fmt.Errorf("error setting canvas background color for zoom: %w", err)
	}

	err = canvas.SetFormat("png")
	if err != nil {
		return nil, fmt.Errorf("error setting canvas format for zoom: %w", err)
	}

	err = canvas.NewImage(originalWidth, originalHeight, imagick.NewPixelWand())
	if err != nil {
		return nil, fmt.Errorf("error creating canvas for zoom: %w", err)
	}

	err = canvas.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_ACTIVATE)
	if err != nil {
		return nil, fmt.Errorf("error setting canvas alpha channel for zoom: %w", err)
	}

	err = canvas.CompositeImage(wand, imagick.COMPOSITE_OP_OVER, false, int((originalWidth-wand.GetImageWidth())/2), int((originalHeight-wand.GetImageHeight())/2))
	if err != nil {
		return nil, fmt.Errorf("error compositing image onto canvas for zoom: %w", err)
	}

	// Flip alpha of the canvas to create a mask for outpainting the edges
	mask := canvas.Clone()
	err = mask.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error negating mask for zoom: %w", err)
	}

	editedImage, err := editImage(canvas, ImageEditArgs{
		Prompt: args.Prompt,
	}, metadata, mask)
	if err != nil {
		return nil, err
	}

	return []*imagick.MagickWand{editedImage}, nil
}
