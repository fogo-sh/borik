package bot

import (
	"errors"
	"fmt"
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

func getStickerUrl(sticker *discordgo.Sticker) (string, error) {
	switch sticker.FormatType {
	case discordgo.StickerFormatTypePNG:
		return fmt.Sprintf(
			"https://media.discordapp.net/stickers/%s.webp?size=1024",
			sticker.ID,
		), nil
	case discordgo.StickerFormatTypeAPNG:
		return fmt.Sprintf(
			"https://media.discordapp.net/stickers/%s.png?size=1024",
			sticker.ID,
		), nil
	case discordgo.StickerFormatTypeLottie:
		return "", errors.New("this command does not currently support Lottie (built-in) stickers")
	default:
		return "", errors.New("unknown sticker format")
	}
}

func Sticker(message *discordgo.MessageCreate, args struct{}) {
	var targetSticker *discordgo.Sticker
	if len(message.StickerItems) >= 1 {
		targetSticker = message.StickerItems[0]
	} else if message.ReferencedMessage != nil && len(message.ReferencedMessage.StickerItems) >= 1 {
		targetSticker = message.ReferencedMessage.StickerItems[0]
	} else {
		messages, err := Instance.Session.ChannelMessages(message.ChannelID, 20, message.ID, "", "")
		if err != nil {
			log.Error().Err(err).Msg("Error fetching message history")
			return
		}

		for _, message := range messages {
			if len(message.StickerItems) >= 1 {
				targetSticker = message.StickerItems[0]
				break
			}
		}
	}

	if targetSticker == nil {
		Instance.Session.ChannelMessageSendReply(
			message.ChannelID,
			"No sticker found! Please post the sticker you are looking for and try again, or retry this command as a reply on the target message.",
			message.Reference(),
		)
		return
	}

	stickerUrl, err := getStickerUrl(targetSticker)
	if err != nil {
		Instance.Session.ChannelMessageSendReply(
			message.ChannelID,
			fmt.Sprintf("Unable to fetch sticker: %s", err),
			message.Reference(),
		)
		return
	}

	resp, err := http.Get(stickerUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading sticker")
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
