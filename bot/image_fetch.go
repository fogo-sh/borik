package bot

import (
	"net/http"
	"path"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type AvatarArgs struct {
	User           string `default:"" description:"User to fetch the avatar for. Must be a single Discord mention."`
	UseGuildAvatar bool   `default:"true" description:"Attempt to fetch the user's guild avatar first. Disable to always use their global avatar.'"`
}

// Avatar fetches a user's avatar.
func Avatar(message *discordgo.MessageCreate, args AvatarArgs) {
	if len(message.Mentions) != 1 {
		Instance.Session.ChannelMessageSendReply(
			message.ChannelID,
			"You must provide a single user to fetch an avatar for, as a Discord mention.",
			message.Reference(),
		)
		return
	}

	targetUser := message.Mentions[0]
	member, err := Instance.Session.GuildMember(message.GuildID, targetUser.ID)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching member")
		return
	}

	var avatarUrl string
	if args.UseGuildAvatar {
		avatarUrl = member.AvatarURL("1024")
	} else {
		avatarUrl = targetUser.AvatarURL("1024")
	}

	resp, err := http.Get(avatarUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading avatar")
		return
	}
	defer resp.Body.Close()

	Instance.Session.ChannelMessageSendComplex(
		message.ChannelID,
		&discordgo.MessageSend{
			Reference: message.Reference(),
			File: &discordgo.File{
				Name:        path.Base(resp.Request.URL.Path),
				ContentType: resp.Header.Get("Content-Type"),
				Reader:      resp.Body,
			},
		},
	)
}
