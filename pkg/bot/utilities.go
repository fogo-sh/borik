package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type ImageOperationArgs interface {
	GetImageURL() string
}

type ImageOperation[K ImageOperationArgs] func(*imagick.MagickWand, K) ([]*imagick.MagickWand, error)

// OperationContext is a unifying struct that abstracts either a MessageCreate or an InteractionCreate.
// It can be constructed with either type, and offers utility methods (such as retrieving a channelID)
// that apply in both contexts, as well as a callback mechanism for context-specific logic.
type OperationContext struct {
	Session     *discordgo.Session
	Message     *discordgo.MessageCreate
	Interaction *discordgo.InteractionCreate
}

// NewOperationContextFromMessage constructs a context for a normal MessageCreate flow.
func NewOperationContextFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) *OperationContext {
	return &OperationContext{
		Session: session,
		Message: message,
	}
}

// NewOperationContextFromInteraction constructs a context for an InteractionCreate (slash command) flow.
func NewOperationContextFromInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) *OperationContext {
	return &OperationContext{
		Session:     session,
		Interaction: interaction,
	}
}

// GetChannelID provides a unified way to retrieve the channel ID for either a message or an interaction.
func (ctx *OperationContext) GetChannelID() string {
	if ctx.Message != nil {
		return ctx.Message.ChannelID
	} else if ctx.Interaction != nil {
		// In slash commands, ChannelID is attached to the Interaction struct
		return ctx.Interaction.ChannelID
	}
	return ""
}

// Reference returns a *discordgo.MessageReference for a message context, or nil for interactions
// (though you can extend this if you want to handle replying differently in an interaction context).
func (ctx *OperationContext) Reference() *discordgo.MessageReference {
	if ctx.Message == nil {
		return nil
	}
	return ctx.Message.Reference()
}

// RunCallbacks allows you to separately handle the two different context types immediately.
// If you pass in both callbacks, only the one matching the current context will be invoked.
func (ctx *OperationContext) RunCallbacks(
	onMessageCreate func(*discordgo.MessageCreate),
	onInteractionCreate func(*discordgo.InteractionCreate),
) {
	if ctx.Message != nil && onMessageCreate != nil {
		onMessageCreate(ctx.Message)
	} else if ctx.Interaction != nil && onInteractionCreate != nil {
		onInteractionCreate(ctx.Interaction)
	}
}

// TypingIndicatorForContext triggers a typing indicator if we are working with a message context,
// and does a no-op for slash commands (or could be customized, e.g., ephemeral or different behaviors).
func TypingIndicatorForContext(ctx *OperationContext) func() {
	if ctx.Message != nil {
		return TypingIndicator(ctx.Message)
	}
	return func() {} // no-op for Interaction-based contexts, or adapt if needed
}

// TypingIndicator invokes a typing indicator in the channel of a message
func TypingIndicator(message *discordgo.MessageCreate) func() {
	stopTyping := Schedule(
		func() {
			log.Debug().Str("channel", message.ChannelID).Msg("Invoking typing indicator in channel")
			err := Instance.session.ChannelTyping(message.ChannelID)
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

// FindImageURLFromMessage attempts to find an image in a given message, falling back to scanning message history if one cannot be found.
func FindImageURLFromMessage(m *discordgo.MessageCreate) (string, error) {
	if imageUrl := ImageURLFromMessage(m.Message); imageUrl != "" {
		return imageUrl, nil
	}

	if m.ReferencedMessage != nil {
		if imageUrl := ImageURLFromMessage(m.ReferencedMessage); imageUrl != "" {
			return imageUrl, nil
		}
	}

	return FindImageURLInChannel(Instance.session, m.ChannelID, m.ID)
}

// FindImageURLInChannel attempts to find an image in a given channel, scanning message history.
func FindImageURLInChannel(s *discordgo.Session, channelID string, beforeID string) (string, error) {
	messages, err := s.ChannelMessages(channelID, 20, beforeID, "", "")
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

// MakeImageOpTextCommand automatically creates a Parsley command handler for a given ImageOperation
func MakeImageOpTextCommand[K ImageOperationArgs](operation ImageOperation[K]) func(*discordgo.MessageCreate, K) {
	return func(message *discordgo.MessageCreate, args K) {
		PrepareAndInvokeOperation(NewOperationContextFromMessage(Instance.session, message), args, operation)
	}
}

func MakeImageOpSlashCommand[K ImageOperationArgs](operation ImageOperation[K]) func(*discordgo.Session, *discordgo.InteractionCreate, K) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, args K) {
		PrepareAndInvokeOperation(NewOperationContextFromInteraction(session, interaction), args, operation)
	}
}

func PrepareAndInvokeOperation[K ImageOperationArgs](ctx *OperationContext, args K, operation ImageOperation[K]) {
	defer TypingIndicatorForContext(ctx)()

	imageUrl := args.GetImageURL()
	if imageUrl == "" {
		var err error

		ctx.RunCallbacks(
			func(m *discordgo.MessageCreate) {
				imageUrl, err = FindImageURLFromMessage(m)
			},
			func(i *discordgo.InteractionCreate) {
				imageUrl, err = FindImageURLInChannel(Instance.session, i.ChannelID, "")
			},
		)

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

	parsedUrl, _ := url.Parse(imageUrl)
	filename := path.Base(parsedUrl.Path)

	input := imagick.NewMagickWand()
	err = input.SetFilename(filename)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set image filename - loading may not behave as expected.")
	}
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

	ctx.RunCallbacks(
		func(m *discordgo.MessageCreate) {
			_, err = ctx.Session.ChannelMessageSendComplex(
				m.ChannelID,
				&discordgo.MessageSend{
					Reference: m.Reference(),
					File: &discordgo.File{
						Name:   fmt.Sprintf("output.%s", strings.ToLower(resultImage.GetImageFormat())),
						Reader: destBuffer,
					},
				},
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to send image")
				_, err = ctx.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
				if err != nil {
					log.Error().Err(err).Msg("Failed to send error message")
				}
			}
		},
		func(i *discordgo.InteractionCreate) {
			err = ctx.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:   fmt.Sprintf("output.%s", strings.ToLower(resultImage.GetImageFormat())),
							Reader: destBuffer,
						},
					},
				},
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send interaction response")
			}
		},
	)
}

// ResizeMaintainAspectRatio resizes an input wand to fit within a box of given width and height, maintaining aspect ratio
func ResizeMaintainAspectRatio(wand *imagick.MagickWand, width uint, height uint) error {
	inputHeight := float64(wand.GetImageHeight())
	inputWidth := float64(wand.GetImageWidth())

	widthMagFactor := float64(width) / inputWidth
	heightMagFactor := float64(height) / inputHeight

	minFactor := math.Min(widthMagFactor, heightMagFactor)

	targetWidth := inputWidth * minFactor
	targetHeight := inputHeight * minFactor

	return wand.ScaleImage(uint(targetWidth), uint(targetHeight))
}

type OverlayOptions struct {
	HFlip bool
	VFlip bool

	OverlayWidthFactor  float64
	OverlayHeightFactor float64

	RightToLeft bool
}

// OverlayImage overlays an image onto another image
func OverlayImage(wand *imagick.MagickWand, overlay []byte, options OverlayOptions) error {
	overlayWand := imagick.NewMagickWand()
	err := overlayWand.ReadImageBlob(overlay)
	if err != nil {
		return fmt.Errorf("error reading overlay: %w", err)
	}

	if options.HFlip {
		err = overlayWand.FlopImage()
		if err != nil {
			return fmt.Errorf("error flipping overlay horizontally: %w", err)
		}
	}
	if options.VFlip {
		err = overlayWand.FlipImage()
		if err != nil {
			return fmt.Errorf("error flipping overlay vertically: %w", err)
		}
	}

	inputWidth := wand.GetImageWidth()
	inputHeight := wand.GetImageHeight()

	err = ResizeMaintainAspectRatio(
		overlayWand,
		uint(float64(inputWidth)*options.OverlayWidthFactor),
		uint(float64(inputHeight)*options.OverlayHeightFactor),
	)
	if err != nil {
		return fmt.Errorf("error resizing overlay: %w", err)
	}

	overlayWidth := overlayWand.GetImageWidth()
	overlayHeight := overlayWand.GetImageHeight()

	if options.HFlip {
		options.RightToLeft = !options.RightToLeft
	}

	xOffset := 0
	if options.RightToLeft {
		xOffset = int(inputWidth - overlayWidth)
	}

	yOffset := 0
	if !options.VFlip {
		yOffset = int(inputHeight - overlayHeight)
	}

	return wand.CompositeImage(overlayWand, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
}

type OverlayImageArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (args OverlayImageArgs) GetImageURL() string {
	return args.ImageURL
}

func MakeImageOverlayCommand(overlayImage []byte, initialOptions OverlayOptions) func(*discordgo.MessageCreate, OverlayImageArgs) {
	return MakeImageOpTextCommand(func(wand *imagick.MagickWand, args OverlayImageArgs) ([]*imagick.MagickWand, error) {
		newOptions := initialOptions

		if args.HFlip {
			newOptions.HFlip = !newOptions.HFlip
		}
		if args.VFlip {
			newOptions.VFlip = !newOptions.VFlip
		}

		err := OverlayImage(
			wand,
			overlayImage,
			newOptions,
		)

		return []*imagick.MagickWand{wand}, err
	})
}
