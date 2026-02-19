package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type HelpArgs struct {
	Command string `default:"" description:"Command to get help for."`
}

func generateCommandList() string {
	commandCodeBlock := "```"

	for _, details := range Instance.textParser.GetCommands() {
		commandCodeBlock += fmt.Sprintf("%s%s: %s\n", Instance.config.Prefix, details.Name, details.Description)
	}

	return commandCodeBlock + "```"
}

func generateCommandHelp(command string) (*discordgo.MessageEmbed, error) {
	commandDetails, err := Instance.textParser.GetCommand(command)
	if err != nil {
		return nil, fmt.Errorf("error getting command details: %w", err)
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s%s", Instance.config.Prefix, command),
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

func help(ctx *OperationContext, args HelpArgs) {
	if args.Command != "" {
		embed, err := generateCommandHelp(args.Command)
		if err != nil {
			if sendErr := ctx.SendText(fmt.Sprintf("```\n%s\n```", err.Error())); sendErr != nil {
				log.Error().Err(sendErr).Msg("Error sending error message")
			}
			return
		}

		if err := ctx.SendEmbed(embed); err != nil {
			log.Error().Err(err).Msg("Failed to send help message")
		}
	} else {
		if err := ctx.SendText(generateCommandList()); err != nil {
			log.Error().Err(err).Msg("Failed to send help message")
		}
	}
}

func HelpCommand(message *discordgo.MessageCreate, args HelpArgs) {
	help(NewOperationContextFromMessage(Instance.session, message), args)
}

func HelpSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args HelpArgs) {
	help(NewOperationContextFromInteraction(session, interaction), args)
}
