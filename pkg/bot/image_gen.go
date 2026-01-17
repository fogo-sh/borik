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
			Content: fmt.Sprintf("Generated image with seed: %d", seed),
			Files: []*discordgo.File{
				{
					Name:        "generated.png",
					ContentType: "image/png",
					Reader:      base64.NewDecoder(base64.StdEncoding, strings.NewReader(image.Data[0].B64JSON)),
				},
			},
			Reference: message.Reference(),
		},
	)
}

type ImageEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
}

func (args ImageEditArgs) GetImageURL() string {
	return args.ImageURL
}

func ImageEdit(wand *imagick.MagickWand, args ImageEditArgs) ([]*imagick.MagickWand, error) {
	imageBlob, err := wand.GetImageBlob()
	if err != nil {
		return nil, fmt.Errorf("error getting image blob: %w", err)
	}
	imageReader := bytes.NewReader(imageBlob)

	seed := rand.Int()
	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	editedImage, err := Instance.openAiClient.Images.Edit(
		context.TODO(),
		openai.ImageEditParams{
			Image: openai.ImageEditParamsImageUnion{
				OfFileArray: []io.Reader{imageReader},
			},
			Prompt:         finalPrompt,
			Model:          "flux-2-klein-4b",
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

	wand = imagick.NewMagickWand()
	err = wand.ReadImageBlob(decodedImg)
	if err != nil {
		return nil, fmt.Errorf("error reading edited image blob: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
