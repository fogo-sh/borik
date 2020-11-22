package bot

import (
	"github.com/bwmarrin/discordgo"
)

type _IMagikArgs struct {
	ImageURL string `default:""`
}

func _IMagikCommand(message *discordgo.MessageCreate, args _IMagikArgs) {
	magickArgs := _MagikArgs{ImageURL: args.ImageURL, Scale: -1}
	_MagikCommand(message, magickArgs)
}
