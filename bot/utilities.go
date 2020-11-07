package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// ImageURLFromMessage attempts to retrieve an image URL from a given message.
func ImageURLFromMessage(m *discordgo.Message) (string, bool) {
	if len(m.Embeds) == 1 {
		embed := m.Embeds[0]

		if embed.Type == "Image" {
			return embed.URL, true
		}
	}

	if len(m.Attachments) == 1 {
		attachment := m.Attachments[0]
		return attachment.URL, true
	}

	return "", false
}

// FindImageURL attempts to find an image in a given message, falling back to scanning message history if one cannot be found.
func FindImageURL(m *discordgo.MessageCreate) (string, error) {
	if embedURL, found := ImageURLFromMessage(m.Message); found {
		return embedURL, nil
	}

	messages, err := Instance.Session.ChannelMessages(m.ChannelID, 20, m.ID, "", "")
	if err != nil {
		return "", fmt.Errorf("error retrieving message history: %w", err)
	}

	for _, message := range messages {
		if embedURL, found := ImageURLFromMessage(message); found {
			return embedURL, nil
		}
	}
	return "", errors.New("unable to locate an image")
}

// DownloadImage downloads an image from a given URL, returing the resulting bytes.
func DownloadImage(url string) ([]byte, error) {
	log.Debug().Str("url", url).Msg("Downloading image")
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading image: %w", err)
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)

	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error copying image to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}
