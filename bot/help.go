package bot

// type _ArgDetails struct {
// 	Name     string
// 	Type     string
// 	Default  string
// 	Required bool
// }

// func _GetArgDetails(command Command) []_ArgDetails {
// 	funcType := reflect.TypeOf(command.Handler)
// 	argsType := funcType.In(1)

// 	args := []_ArgDetails{}

// 	for index := 0; index < argsType.NumField(); index++ {
// 		arg := argsType.Field(index)

// 		defaultVal, hasDefault := arg.Tag.Lookup("default")
// 		args = append(args, _ArgDetails{arg.Name, arg.Type.Name(), defaultVal, !hasDefault})
// 	}
// 	return args
// }

// func _HelpCommand(message *discordgo.MessageCreate, args struct{}) {
// 	embed := &discordgo.MessageEmbed{
// 		Title:  "Borik Help",
// 		Fields: []*discordgo.MessageEmbedField{},
// 		Color:  (206 << 16) + (147 << 8) + 216,
// 	}

// 	for command, details := range Instance.Commands {
// 		argString := ""
// 		for _, argDetails := range _GetArgDetails(details) {
// 			if argDetails.Required {
// 				argString += fmt.Sprintf(" <%s:%s>", argDetails.Name, argDetails.Type)
// 			} else {
// 				argString += fmt.Sprintf(" [%s:%s=%s]", argDetails.Name, argDetails.Type, argDetails.Default)
// 			}
// 		}
// 		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
// 			Name:  fmt.Sprintf("%s%s%s", Instance.Config.Prefix, command, argString),
// 			Value: details.Description,
// 		})
// 	}

// 	_, err := Instance.Session.ChannelMessageSendEmbed(message.ChannelID, embed)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to send help message")
// 	}
// }
