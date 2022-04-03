package bot

import (
	"github.com/bwmarrin/discordgo"
)

type _MagikArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale    float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
}

func (args _MagikArgs) GetImageURL() string {
	return args.ImageURL
}

func _MagikCommand(message *discordgo.MessageCreate, args _MagikArgs) {
	defer TypingIndicator(message)()

	PrepareAndInvokeOperation(message, args, Magik)
}
