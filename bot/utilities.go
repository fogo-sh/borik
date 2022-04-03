package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type ImageOperationArgs interface {
	GetImageURL() string
}

// TypingIndicator invokes a typing indicator in the channel of a message
func TypingIndicator(message *discordgo.MessageCreate) func() {
	stopTyping := Schedule(
		func() {
			log.Debug().Str("channel", message.ChannelID).Msg("Invoking typing indicator in channel")
			err := Instance.Session.ChannelTyping(message.ChannelID)
			if err != nil {
				log.Error().Err(err).Msg("Error while attempting invoke typing indicator in channel")
				return
			}
		},
		5*time.Second,
	)
	return func() {
		stopTyping <- true
	}
}

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

// Schedule some func to be ran in a cancelable goroutine on an interval
func Schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// PrepareAndInvokeOperation downloads the image pulled from the message, invokes the given operation with said image, and posts the image in the channel of the message that invoked it
func PrepareAndInvokeOperation[K ImageOperationArgs](message *discordgo.MessageCreate, args K, operation func([]byte, io.Writer, K) error) {
	imageUrl := args.GetImageURL()
	if imageUrl == "" {
		var err error
		imageUrl, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	srcBytes, err := DownloadImage(imageUrl)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}
	destBuffer := new(bytes.Buffer)

	log.Debug().Msg("Beginning processing image")
	err = operation(srcBytes, destBuffer, args)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process image")
		return
	}

	log.Debug().Msg("Image processed, uploading result")
	_, err = Instance.Session.ChannelFileSend(message.ChannelID, "test.jpeg", destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		_, err = Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
	}
}
