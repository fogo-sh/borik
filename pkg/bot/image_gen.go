package bot

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go/v3"
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
			Model:          "z-image-turbo",
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
