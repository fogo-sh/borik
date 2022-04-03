package bot

import (
	"github.com/bwmarrin/discordgo"
)

type _ArcweldArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args _ArcweldArgs) GetImageURL() string {
	return args.ImageURL
}

func _ArcweldCommand(message *discordgo.MessageCreate, args _ArcweldArgs) {
	PrepareAndInvokeOperation(message, args, Arcweld)
}
