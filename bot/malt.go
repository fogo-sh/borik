package bot

import (
	"github.com/bwmarrin/discordgo"
)

type _MaltArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degree   float64 `default:"45" description:"Number of degrees to rotate the image by while processing."`
}

func (args _MaltArgs) GetImageURL() string {
	return args.ImageURL
}

func _MaltCommand(message *discordgo.MessageCreate, args _MaltArgs) {
	defer TypingIndicator(message)()

	PrepareAndInvokeOperation(message, args, Malt)
}
