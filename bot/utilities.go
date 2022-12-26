package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v2/imagick"
)

type ImageOperationArgs interface {
	GetImageURL() string
}

type ImageOperation[K ImageOperationArgs] func(*imagick.MagickWand, K) ([]*imagick.MagickWand, error)

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

// Schedule some func to be run in a cancelable goroutine on an interval
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

// ImageURLFromMessage attempts to retrieve an image URL from a given message.
func ImageURLFromMessage(m *discordgo.Message) string {
	for _, embed := range m.Embeds {
		if embed.Type == "Image" {
			return embed.URL
		}
	}

	for _, attachment := range m.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image/") {
			return attachment.URL
		}
	}

	return ""
}

// FindImageURL attempts to find an image in a given message, falling back to scanning message history if one cannot be found.
func FindImageURL(m *discordgo.MessageCreate) (string, error) {
	if imageUrl := ImageURLFromMessage(m.Message); imageUrl != "" {
		return imageUrl, nil
	}

	messages, err := Instance.Session.ChannelMessages(m.ChannelID, 20, m.ID, "", "")
	if err != nil {
		return "", fmt.Errorf("error retrieving message history: %w", err)
	}

	for _, message := range messages {
		if imageUrl := ImageURLFromMessage(message); imageUrl != "" {
			return imageUrl, nil
		}
	}
	return "", errors.New("unable to locate an image")
}

// DownloadImage downloads an image from a given URL, returning the resulting bytes.
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

// MakeImageOpCommand automatically creates a Parsley command handler for a given ImageOperation
func MakeImageOpCommand[K ImageOperationArgs](operation ImageOperation[K]) func(*discordgo.MessageCreate, K) {
	return func(message *discordgo.MessageCreate, args K) {
		PrepareAndInvokeOperation(message, args, operation)
	}
}

// PrepareAndInvokeOperation automatically handles invoking a given ImageOperation for a Discord message and returning the finished results.
func PrepareAndInvokeOperation[K ImageOperationArgs](message *discordgo.MessageCreate, args K, operation ImageOperation[K]) {
	defer TypingIndicator(message)()

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

	input := imagick.NewMagickWand()
	err = input.ReadImageBlob(srcBytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read image")
		return
	}
	input = input.CoalesceImages()

	var resultFrames []*imagick.MagickWand

	for i := 0; i < int(input.GetNumberImages()); i++ {
		input.SetIteratorIndex(i)
		inputFrame := input.GetImage().Clone()
		log.Debug().Int("frame", i).Msg("Beginning processing frame")
		output, err := operation(inputFrame, args)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process image")
			return
		}
		resultFrames = append(resultFrames, output...)
	}

	resultImage := imagick.NewMagickWand()

	for index, frame := range resultFrames {
		log.Debug().Int("frame", index).Msg("Adding frame to result image")
		err := resultImage.AddImage(frame)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add frame")
			return
		}
	}

	input.ResetIterator()
	resultImage.ResetIterator()

	log.Debug().Msg("Setting image format")
	if len(resultFrames) > 1 {
		err := resultImage.SetImageFormat("GIF")
		if err != nil {
			log.Error().Err(err).Msg("Failed to set result format")
			return
		}
		err = resultImage.SetImageDelay(input.GetImageDelay())
		if err != nil {
			log.Error().Err(err).Msg("Failed to set framerate")
			return
		}
	} else {
		err := resultImage.SetImageFormat("PNG")
		if err != nil {
			log.Error().Err(err).Msg("Failed to set result format")
			return
		}
	}

	log.Debug().Msg("Repaging image")
	err = resultImage.ResetImagePage("0x0+0+0")
	if err != nil {
		log.Error().Err(err).Msg("Failed to repage image")
	}

	log.Debug().Msg("Deconstructing image")
	resultImage = resultImage.DeconstructImages()
	destBuffer := new(bytes.Buffer)

	log.Debug().Msg("Writing output image")
	_, err = destBuffer.Write(resultImage.GetImagesBlob())
	if err != nil {
		log.Error().Err(err).Msg("Failed to write image")
		return
	}

	log.Debug().Msg("Image processed, uploading result")
	_, err = Instance.Session.ChannelFileSend(
		message.ChannelID,
		fmt.Sprintf("output.%s", strings.ToLower(resultImage.GetImageFormat())),
		destBuffer,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		_, err = Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
	}
}
