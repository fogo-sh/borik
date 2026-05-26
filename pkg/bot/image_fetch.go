package bot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	jobArgs "github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/utils"
)

type AvatarArgs struct {
	User           string `default:"" description:"User to fetch the avatar for. Must be a single Discord mention."`
	UseGuildAvatar bool   `default:"true" description:"Attempt to fetch the user's guild avatar first. Disable to always use their global avatar.'"`
}

type AvatarSlashArgs struct {
	User           *discordgo.User `description:"User to fetch the avatar for. Defaults to yourself."`
	UseGuildAvatar bool            `default:"true" description:"Attempt to fetch the user's guild avatar first. Disable to always use their global avatar."`
}

func fetchAvatar(ctx *OperationContext, targetUser *discordgo.User, guildID string, useGuildAvatar bool) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	member, err := ctx.Session.GuildMember(guildID, targetUser.ID)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching member")
		return
	}

	var avatarUrl string
	if useGuildAvatar {
		avatarUrl = member.AvatarURL("1024")
	} else {
		avatarUrl = targetUser.AvatarURL("1024")
	}

	resp, err := http.Get(avatarUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading avatar")
		return
	}
	defer utils.CloseBody(resp.Body, "Error closing avatar response body")

	file := &discordgo.File{
		Name:        path.Base(resp.Request.URL.Path),
		ContentType: resp.Header.Get("Content-Type"),
		Reader:      resp.Body,
	}

	if err := ctx.SendFiles([]*discordgo.File{file}); err != nil {
		log.Error().Err(err).Msg("Failed to send avatar")
	}
}

// Avatar fetches a user's avatar.
func Avatar(message *discordgo.MessageCreate, args AvatarArgs) {
	if len(message.Mentions) != 1 {
		_, err := Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"You must provide a single user to fetch an avatar for, as a Discord mention.",
			message.Reference(),
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send avatar usage error")
		}
		return
	}

	fetchAvatar(
		NewOperationContextFromMessage(Instance.session, message),
		message.Mentions[0],
		message.GuildID,
		args.UseGuildAvatar,
	)
}

func AvatarSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args AvatarSlashArgs) {
	ctx := NewOperationContextFromInteraction(session, interaction)

	targetUser := args.User
	if targetUser == nil {
		if interaction.Member != nil {
			targetUser = interaction.Member.User
		} else if interaction.User != nil {
			targetUser = interaction.User
		}
	}

	if targetUser == nil {
		if err := ctx.SendText("Unable to determine target user."); err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
		return
	}

	fetchAvatar(ctx, targetUser, interaction.GuildID, args.UseGuildAvatar)
}

func getStickerUrl(sticker *discordgo.StickerItem) (string, string, error) {
	switch sticker.FormatType {
	case discordgo.StickerFormatTypePNG:
		return fmt.Sprintf(
			"https://media.discordapp.net/stickers/%s.webp?size=1024",
			sticker.ID,
		), "image/webp", nil
	case discordgo.StickerFormatTypeAPNG:
		return fmt.Sprintf(
			"https://media.discordapp.net/stickers/%s.png?size=1024",
			sticker.ID,
		), "image/gif", nil
	case discordgo.StickerFormatTypeGIF:
		return fmt.Sprintf(
			"https://media.discordapp.net/stickers/%s.gif?size=1024",
			sticker.ID,
		), "image/gif", nil
	case discordgo.StickerFormatTypeLottie:
		return "", "", errors.New("this command does not currently support Lottie / built-in stickers")
	default:
		return "", "", errors.New("unknown sticker format")
	}
}

func Sticker(message *discordgo.MessageCreate, args struct{}) {
	var targetSticker *discordgo.StickerItem
	if len(message.StickerItems) >= 1 {
		targetSticker = message.StickerItems[0]
	} else if message.ReferencedMessage != nil && len(message.ReferencedMessage.StickerItems) >= 1 {
		targetSticker = message.ReferencedMessage.StickerItems[0]
	} else {
		messages, err := Instance.session.ChannelMessages(message.ChannelID, 20, message.ID, "", "")
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
		_, err := Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"No sticker found! Please post the sticker you are looking for and try again, "+
				"or retry this command as a reply on the target message.",
			message.Reference(),
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send missing sticker error")
		}
		return
	}

	stickerUrl, contentType, err := getStickerUrl(targetSticker)
	if err != nil {
		_, err := Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			fmt.Sprintf("Unable to fetch sticker: %s", err),
			message.Reference(),
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send sticker fetch error")
		}
		return
	}

	var file io.Reader
	var filename string
	if targetSticker.FormatType == discordgo.StickerFormatTypeAPNG {
		parsedURL, _ := url.Parse(stickerUrl)
		target := NewOperationContextFromMessage(Instance.session, message).DeliveryTarget("Error converting APNG sticker to GIF")
		target.Filename = path.Base(parsedURL.Path) + ".gif"
		target.ContentType = contentType

		err = Instance.triggerAPNGToGIF(
			context.Background(),
			message.ID+"-apng-to-gif",
			jobArgs.APNGToGIF{ImageURL: stickerUrl},
			target,
		)
		if err != nil {
			_, sendErr := Instance.session.ChannelMessageSendReply(
				message.ChannelID,
				fmt.Sprintf("Error starting APNG sticker conversion:\n```%s```", err),
				message.Reference(),
			)
			if sendErr != nil {
				log.Error().Err(sendErr).Msg("Failed to send sticker conversion error")
			}
			log.Error().Err(err).Msg("Error starting APNG sticker conversion")
			return
		}
		return
	} else {
		resp, err := http.Get(stickerUrl)
		if err != nil {
			log.Error().Err(err).Msg("Error downloading sticker")
			return
		}
		defer utils.CloseBody(resp.Body, "Error closing sticker response body")

		file = resp.Body
		filename = path.Base(resp.Request.URL.Path)
	}

	_, err = Instance.session.ChannelMessageSendComplex(
		message.ChannelID,
		&discordgo.MessageSend{
			Reference: message.Reference(),
			File: &discordgo.File{
				Name:        filename,
				ContentType: contentType,
				Reader:      file,
			},
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send sticker")
	}
}

func getEmojiUrl(emoji *discordgo.Emoji) string {
	if emoji.Animated {
		return fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.gif?size=1024&quality=lossless", emoji.ID)
	} else {
		return fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.webp?size=1024&quality=lossless", emoji.ID)
	}
}

type EmojiArgs struct {
	Emoji string `description:"Emoji to fetch as an image. Leave blank to attempt to auto-locate an emoji." default:""`
}

func Emoji(message *discordgo.MessageCreate, args EmojiArgs) {
	var targetEmoji *discordgo.Emoji
	if len(message.GetCustomEmojis()) >= 1 {
		targetEmoji = message.GetCustomEmojis()[0]
	} else if message.ReferencedMessage != nil && len(message.ReferencedMessage.GetCustomEmojis()) >= 1 {
		targetEmoji = message.ReferencedMessage.GetCustomEmojis()[0]
	} else {
		messages, err := Instance.session.ChannelMessages(message.ChannelID, 20, message.ID, "", "")
		if err != nil {
			log.Error().Err(err).Msg("Error fetching message history")
			return
		}

		for _, message := range messages {
			if len(message.GetCustomEmojis()) >= 1 {
				targetEmoji = message.GetCustomEmojis()[0]
				break
			}
		}
	}

	if targetEmoji == nil {
		_, err := Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"No emoji found! Please post the emoji you are looking for and try again, "+
				"or retry this command as a reply on the target message.",
			message.Reference(),
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send missing emoji error")
		}
		return
	}

	emojiUrl := getEmojiUrl(targetEmoji)

	resp, err := http.Get(emojiUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading emoji")
		return
	}
	defer utils.CloseBody(resp.Body, "Error closing emoji response body")

	_, err = Instance.session.ChannelMessageSendComplex(
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
	if err != nil {
		log.Error().Err(err).Msg("Failed to send emoji")
	}
}
