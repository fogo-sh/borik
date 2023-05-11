package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/rs/zerolog/log"

	"github.com/fogo-sh/borik/pkg/config"
)

func generateCommandList(parser *parsley.Parser) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:  "Borik Help",
		Fields: []*discordgo.MessageEmbedField{},
		Color:  (206 << 16) + (147 << 8) + 216,
	}

	for _, details := range parser.GetCommands() {
		argString := ""
		for _, argDetails := range details.Arguments {
			if argDetails.Required {
				argString += fmt.Sprintf(" <%s:%s>", argDetails.Name, argDetails.Type)
			} else {
				argString += fmt.Sprintf(" [%s:%s=%s]", argDetails.Name, argDetails.Type, argDetails.Default)
			}
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s%s%s", config.Instance.Prefix, details.Name, argString),
			Value: details.Description,
		})
	}
	return embed
}

func generateCommandHelp(parser *parsley.Parser, command string) (*discordgo.MessageEmbed, error) {
	commandDetails, err := parser.GetCommand(command)
	if err != nil {
		return nil, fmt.Errorf("error getting command details: %w", err)
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s%s", config.Instance.Prefix, command),
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

type HelpArgs struct {
	Command string `default:"" description:"Command to get help for."`
}

func (b *Bot) helpCommand(message *discordgo.MessageCreate, args HelpArgs) {
	var embed *discordgo.MessageEmbed
	if args.Command != "" {
		var err error
		embed, err = generateCommandHelp(b.parser, args.Command)
		if err != nil {
			_, err := b.session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\n%s\n```", err.Error()))
			if err != nil {
				log.Error().Err(err).Msg("Error sending error message")
			}
			return
		}

	} else {
		embed = generateCommandList(b.parser)
	}

	_, err := b.session.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send help message")
	}
}
