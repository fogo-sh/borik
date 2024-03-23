package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type HelpArgs struct {
	Command string `default:"" description:"Command to get help for."`
}

func generateCommandList() *discordgo.MessageEmbed {
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
	return embed
}

func generateCommandHelp(command string) (*discordgo.MessageEmbed, error) {
	commandDetails, err := Instance.Parser.GetCommand(command)
	if err != nil {
		return nil, fmt.Errorf("error getting command details: %w", err)
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s%s", Instance.Config.Prefix, command),
		Description: commandDetails.Description,
		Fields:      []*discordgo.MessageEmbedField{},
		Color:       (206 << 16) + (147 << 8) + 216,
	}

	for _, argDetails := range commandDetails.Arguments {
		argDetailsStr := ""
		if argDetails.Required {
			argDetailsStr = fmt.Sprintf("%s\nType: %s", argDetails.Description, argDetails.Type)
		} else {
			argDetailsStr = fmt.Sprintf("%s\nType: %s\nDefault: %s", argDetails.Description, argDetails.Type, argDetails.Default)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  argDetails.Name,
			Value: argDetailsStr,
		})
	}
	return embed, nil
}

func HelpCommand(message *discordgo.MessageCreate, args HelpArgs) {
	var embed *discordgo.MessageEmbed
	if args.Command != "" {
		var err error
		embed, err = generateCommandHelp(args.Command)
		if err != nil {
			_, err := Instance.session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\n%s\n```", err.Error()))
			if err != nil {
				log.Error().Err(err).Msg("Error sending error message")
			}
			return
		}

	} else {
		embed = generateCommandList()
	}

	_, err := Instance.session.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send help message")
	}
}
