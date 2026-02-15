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
	"gopkg.in/gographics/imagick.v3/imagick"
)

type ImageGenArgs struct {
	Prompt string `description:"Prompt to generate an image for."`
}

func ImageGen(message *discordgo.MessageCreate, args ImageGenArgs) {
	defer TypingIndicator(message)()

	seed := rand.Int()
	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	image, err := Instance.openAiClient.Images.Generate(
		context.TODO(),
		openai.ImageGenerateParams{
			Prompt:         finalPrompt,
			Size:           "512x512",
			Model:          "flux-2-klein-4b",
			ResponseFormat: openai.ImageGenerateParamsResponseFormatB64JSON,
		},
	)
	if err != nil {
		Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"Error generating image: `"+err.Error()+"`",
			message.Reference(),
		)
		return
	}

	Instance.session.ChannelMessageSendComplex(
		message.ChannelID,
		&discordgo.MessageSend{
			Reference: message.Reference(),
			Files: []*discordgo.File{
				{
					Name:        "generated.png",
					ContentType: "image/png",
					Reader:      base64.NewDecoder(base64.StdEncoding, strings.NewReader(image.Data[0].B64JSON)),
				},
			},
		},
	)
}

func editImage(wand *imagick.MagickWand, args ImageEditArgs, seed int) (*imagick.MagickWand, error) {
	imageBlob, err := wand.GetImageBlob()
	if err != nil {
		return nil, fmt.Errorf("error getting image blob: %w", err)
	}
	imageReader := bytes.NewReader(imageBlob)

	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	editedImage, err := Instance.openAiClient.Images.Edit(
		context.TODO(),
		openai.ImageEditParams{
			Image: openai.ImageEditParamsImageUnion{
				OfFileArray: []io.Reader{imageReader},
			},
			Prompt: finalPrompt,
			Model:  "flux-2-klein-4b",
			// OpenAI's models require one of a few specific sizes, but stable-diffusion.cpp is more flexible
			// Pass the original image size to prevent it cropping it
			Size: openai.ImageEditParamsSize(fmt.Sprintf(
				"%dx%d",
				wand.GetImageWidth(),
				wand.GetImageHeight(),
			)),
			ResponseFormat: openai.ImageEditParamsResponseFormatB64JSON,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error editing image: %w", err)
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

func ImageEdit(wand *imagick.MagickWand, args ImageEditArgs, seed int) ([]*imagick.MagickWand, error) {
	editedImage, err := editImage(wand, args, seed)
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

func LoopEdit(wand *imagick.MagickWand, args LoopEditArgs, seed int) ([]*imagick.MagickWand, error) {
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps)

	currentWand := wand
	var err error
	for range args.Steps {
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt,
		}, seed)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
}

func ImageEditCommand(message *discordgo.MessageCreate, args ImageEditArgs) {
	seed := rand.Int()
	PrepareAndInvokeOperation(message, args, func(wand *imagick.MagickWand, args ImageEditArgs) ([]*imagick.MagickWand, error) {
		return ImageEdit(wand, args, seed)
	})
}

func LoopEditCommand(message *discordgo.MessageCreate, args LoopEditArgs) {
	seed := rand.Int()
	PrepareAndInvokeOperation(message, args, func(wand *imagick.MagickWand, args LoopEditArgs) ([]*imagick.MagickWand, error) {
		return LoopEdit(wand, args, seed)
	})
}
