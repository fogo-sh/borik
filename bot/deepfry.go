package bot

import (
	"github.com/bwmarrin/discordgo"
)

type _DeepfryArgs struct {
	ImageURL        string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius      float64 `default:"100" description:"Radius of outline to draw around edges."`
	DownscaleFactor uint    `default:"2" description:"Factor to downscale the image by while processing."`
}

func (args _DeepfryArgs) GetImageURL() string {
	return args.ImageURL
}

func _DeepfryCommand(message *discordgo.MessageCreate, args _DeepfryArgs) {
	PrepareAndInvokeOperation(message, args, Deepfry)
}
