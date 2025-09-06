package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type AvatarArgs struct {
	User           string `default:"" description:"User to fetch the avatar for. Must be a single Discord mention."`
	UseGuildAvatar bool   `default:"true" description:"Attempt to fetch the user's guild avatar first. Disable to always use their global avatar.'"`
}

// Avatar fetches a user's avatar.
func Avatar(message *discordgo.MessageCreate, args AvatarArgs) {
	if len(message.Mentions) != 1 {
		Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"You must provide a single user to fetch an avatar for, as a Discord mention.",
			message.Reference(),
		)
		return
	}

	targetUser := message.Mentions[0]
	member, err := Instance.session.GuildMember(message.GuildID, targetUser.ID)
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

	Instance.session.ChannelMessageSendComplex(
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
	case discordgo.StickerFormatTypeLottie:
		return "", "", errors.New("this command does not currently support Lottie / built-in stickers")
	default:
		return "", "", errors.New("unknown sticker format")
	}
}

func apngToGif(apngInput io.Reader) (io.Reader, error) {
	input, err := io.ReadAll(apngInput)
	if err != nil {
		return nil, fmt.Errorf("error copying input: %w", err)
	}

	wand := imagick.NewMagickWand()

	err = wand.SetFilename("APNG:profile.png")
	if err != nil {
		return nil, fmt.Errorf("error setting format: %w", err)
	}

	err = wand.ReadImageBlob(input)
	if err != nil {
		return nil, fmt.Errorf("error reading input image: %w", err)
	}

	for i := uint(0); i < wand.GetNumberImages(); i++ {
		wand.SetIteratorIndex(int(i))
		err = wand.SetImageDispose(imagick.DISPOSE_BACKGROUND)
		if err != nil {
			return nil, fmt.Errorf("error configuring disposal: %w", err)
		}
	}

	wand.ResetIterator()
	wand.CoalesceImages()

	err = wand.SetFilename("profile.gif")
	if err != nil {
		return nil, fmt.Errorf("error setting format: %w", err)
	}

	imageBlob, err := wand.GetImagesBlob()
	if err != nil {
		return nil, fmt.Errorf("error generating output image: %w", err)
	}

	outBuffer := new(bytes.Buffer)
	_, err = outBuffer.Write(imageBlob)
	if err != nil {
		return nil, fmt.Errorf("error outputting image: %w", err)
	}

	if outBuffer.Len() == 0 {
		return nil, fmt.Errorf("got an empty output image - your provided sticker may be one of the currently broken ones")
	}

	return outBuffer, nil
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
		Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"No sticker found! Please post the sticker you are looking for and try again, or retry this command as a reply on the target message.",
			message.Reference(),
		)
		return
	}

	stickerUrl, contentType, err := getStickerUrl(targetSticker)
	if err != nil {
		Instance.session.ChannelMessageSendReply(
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

	var file io.Reader
	var filename string
	if targetSticker.FormatType == discordgo.StickerFormatTypeAPNG {
		file, err = apngToGif(resp.Body)
		if err != nil {
			Instance.session.ChannelMessageSendReply(
				message.ChannelID,
				fmt.Sprintf("Error converting APNG sticker to GIF:\n```%s```", err),
				message.Reference(),
			)
			log.Error().Err(err).Msg("Error converting APNG sticker to GIF")
			return
		}
		filename = path.Base(resp.Request.URL.Path) + ".gif"
	} else {
		file = resp.Body
		filename = path.Base(resp.Request.URL.Path)
	}

	Instance.session.ChannelMessageSendComplex(
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
		Instance.session.ChannelMessageSendReply(
			message.ChannelID,
			"No emoji found! Please post the emoji you are looking for and try again, or retry this command as a reply on the target message.",
			message.Reference(),
		)
		return
	}

	emojiUrl := getEmojiUrl(targetEmoji)

	resp, err := http.Get(emojiUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading emoji")
		return
	}
	defer resp.Body.Close()

	Instance.session.ChannelMessageSendComplex(
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
