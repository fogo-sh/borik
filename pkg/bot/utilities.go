package bot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// imageURLFromMessage attempts to retrieve an image URL from a given message.
func imageURLFromMessage(m *discordgo.Message) string {
	for _, embed := range m.Embeds {
		if embed.Type == "Image" {
			return embed.URL
		} else if embed.Image != nil {
			return embed.Image.URL
		}
	}

	for _, attachment := range m.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image/") {
			return attachment.URL
		}
	}

	return ""
}

// findImageURL attempts to find an image in a given message, falling back to scanning message history if one cannot be found.
func findImageURL(session *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	if imageUrl := imageURLFromMessage(m.Message); imageUrl != "" {
		return imageUrl, nil
	}

	if m.ReferencedMessage != nil {
		if imageUrl := imageURLFromMessage(m.ReferencedMessage); imageUrl != "" {
			return imageUrl, nil
		}
	}

	messages, err := session.ChannelMessages(m.ChannelID, 20, m.ID, "", "")
	if err != nil {
		return "", fmt.Errorf("error retrieving message history: %w", err)
	}

	for _, message := range messages {
		if imageUrl := imageURLFromMessage(message); imageUrl != "" {
			return imageUrl, nil
		}
	}
	return "", errors.New("unable to locate an image")
}
