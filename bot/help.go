package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func _HelpCommand(message *discordgo.MessageCreate, args struct{}) {
	embed := &discordgo.MessageEmbed{
		Title:  "Borik Help",
		Fields: []*discordgo.MessageEmbedField{},
		Color:  (206 << 16) + (147 << 8) + 216,
	}

	for _, details := range Instance.Parser.GetCommands() {
		argString := ""
		for _, argDetails := range details.Arguments {
			if argDetails.Required {
				argString += fmt.Sprintf(" <%s:%s>", argDetails.Name, argDetails.Type)
			} else {
				argString += fmt.Sprintf(" [%s:%s=%s]", argDetails.Name, argDetails.Type, argDetails.Default)
			}
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s%s%s", Instance.Config.Prefix, details.Name, argString),
			Value: details.Description,
		})
	}

	_, err := Instance.Session.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send help message")
	}
}
