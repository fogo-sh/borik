package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
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

type OperationContext struct {
	Session     *discordgo.Session
	Message     *discordgo.MessageCreate
	Interaction *discordgo.InteractionCreate
	deferred    bool
}

func NewOperationContextFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) *OperationContext {
	return &OperationContext{
		Session: session,
		Message: message,
	}
}

func NewOperationContextFromInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) *OperationContext {
	return &OperationContext{
		Session:     session,
		Interaction: interaction,
	}
}

func (ctx *OperationContext) GetSourceID() string {
	if ctx.Message != nil {
		return ctx.Message.ID
	} else if ctx.Interaction != nil {
		return ctx.Interaction.ID
	}
	return ""
}

func (ctx *OperationContext) GetUserID() string {
	if ctx.Message != nil {
		return ctx.Message.Author.ID
	} else if ctx.Interaction != nil {
		if ctx.Interaction.Member != nil {
			return ctx.Interaction.Member.User.ID
		}
		if ctx.Interaction.User != nil {
			return ctx.Interaction.User.ID
		}
	}
	return ""
}

func (ctx *OperationContext) GetChannelID() string {
	if ctx.Message != nil {
		return ctx.Message.ChannelID
	} else if ctx.Interaction != nil {
		return ctx.Interaction.ChannelID
	}
	panic("OperationContext has neither Message nor Interaction set")
}

// DeferResponse defers the interaction response for long-running operations.
// This is a no-op for message-based commands (typing indicator handles that case).
func (ctx *OperationContext) DeferResponse() error {
	if ctx.Interaction == nil {
		return nil
	}
	err := ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to send deferred interaction response: %w", err)
	}
	ctx.deferred = true
	return nil
}

// SendText sends a plain text message.
func (ctx *OperationContext) SendText(content string) error {
	if ctx.Message != nil {
		_, err := ctx.Session.ChannelMessageSendReply(ctx.Message.ChannelID, content, ctx.Message.Reference())
		return err
	}
	if ctx.deferred {
		_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// SendEmbed sends an embed message.
func (ctx *OperationContext) SendEmbed(embed *discordgo.MessageEmbed) error {
	if ctx.Message != nil {
		_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, embed)
		return err
	}
	if ctx.deferred {
		_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		return err
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// SendFiles sends one or more file attachments.
func (ctx *OperationContext) SendFiles(files []*discordgo.File) error {
	if ctx.Message != nil {
		_, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, &discordgo.MessageSend{
			Reference: ctx.Message.Reference(),
			Files:     files,
		})
		return err
	}
	if ctx.deferred {
		_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Files: files,
		})
		return err
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: files,
		},
	})
}

// FindImageURL attempts to locate an image URL from the context.
func (ctx *OperationContext) FindImageURL() (string, error) {
	if ctx.Message != nil {
		return FindImageURLFromMessage(ctx.Message)
	}
	return FindImageURLInChannel(ctx.Session, ctx.Interaction.ChannelID, "")
}

func TypingIndicatorForContext(ctx *OperationContext) func() {
	if ctx.Message != nil {
		return TypingIndicator(ctx.Message)
	}
	return func() {}
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

// ImageURLFromComponent attempts to retrieve an image URL from a given component, recursing into child components if necessary
func ImageURLFromComponent(component discordgo.MessageComponent) string {
	switch c := component.(type) {
	case *discordgo.MediaGallery:
		if len(c.Items) == 0 {
			return ""
		}

		return c.Items[0].Media.URL
	case *discordgo.Container:
		for _, child := range c.Components {
			url := ImageURLFromComponent(child)
			if url != "" {
				return url
			}
		}

		return ""
	default:
		return ""
	}
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

	for _, component := range m.Components {
		url := ImageURLFromComponent(component)
		if url != "" {
			return url
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

// AIImageOperation is like ImageOperation but also receives AISessionMetadata for session tracking and seed management
type AIImageOperation[K ImageOperationArgs] func(*imagick.MagickWand, K, AISessionMetadata) ([]*imagick.MagickWand, error)

// MakeAIImageOpTextCommand creates a Parsley command handler for an AIImageOperation,
// building full AISessionMetadata (with seed, session ID, user ID) from the OperationContext
func MakeAIImageOpTextCommand[K ImageOperationArgs](operation AIImageOperation[K]) func(*discordgo.MessageCreate, K) {
	return func(message *discordgo.MessageCreate, args K) {
		ctx := NewOperationContextFromMessage(Instance.session, message)
		metadata := AISessionMetadata{
			Seed:      rand.Int(),
			SessionID: ctx.GetSourceID(),
			UserID:    ctx.GetUserID(),
		}
		PrepareAndInvokeOperation(ctx, args, func(wand *imagick.MagickWand, args K) ([]*imagick.MagickWand, error) {
			return operation(wand, args, metadata)
		})
	}
}

// MakeAIImageOpSlashCommand creates a slash command handler for an AIImageOperation,
// building full AISessionMetadata (with seed, session ID, user ID) from the OperationContext
func MakeAIImageOpSlashCommand[K ImageOperationArgs](operation AIImageOperation[K]) func(*discordgo.Session, *discordgo.InteractionCreate, K) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, args K) {
		ctx := NewOperationContextFromInteraction(session, interaction)
		metadata := AISessionMetadata{
			Seed:      rand.Int(),
			SessionID: ctx.GetSourceID(),
			UserID:    ctx.GetUserID(),
		}
		PrepareAndInvokeOperation(ctx, args, func(wand *imagick.MagickWand, args K) ([]*imagick.MagickWand, error) {
			return operation(wand, args, metadata)
		})
	}
}

// PrepareAndInvokeOperation automatically handles invoking a given ImageOperation and returning the finished results
func PrepareAndInvokeOperation[K ImageOperationArgs](ctx *OperationContext, args K, operation ImageOperation[K]) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	imageUrl := args.GetImageURL()
	if imageUrl == "" {
		var err error
		imageUrl, err = ctx.FindImageURL()
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

	imageBlob, err := resultImage.GetImagesBlob()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image blob")
		return
	}

	log.Debug().Msg("Writing output image")
	_, err = destBuffer.Write(imageBlob)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write image")
		return
	}

	originalFileName := path.Base(imageUrl)
	originalFileNameNoExt := strings.TrimSuffix(originalFileName, path.Ext(originalFileName))

	log.Debug().Msg("Image processed, uploading result")

	resultFileName := fmt.Sprintf("%s.%s", originalFileNameNoExt, strings.ToLower(resultImage.GetImageFormat()))
	err = ctx.SendFiles([]*discordgo.File{
		{
			Name:   resultFileName,
			Reader: destBuffer,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to send resulting image: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
	}
}

// FindTransparentOpeningRect finds the bounding rectangle of the transparent region in an image
func FindTransparentOpeningRect(frame *imagick.MagickWand) (x, y, width, height int, err error) {
	analysis := frame.Clone()
	defer analysis.Destroy()

	if err := analysis.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_EXTRACT); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error extracting alpha channel: %w", err)
	}
	if err := analysis.NegateImage(false); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error negating alpha mask: %w", err)
	}
	if err := analysis.TrimImage(0); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error trimming: %w", err)
	}

	_, _, ox, oy, err := analysis.GetImagePage()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error getting trimmed region geometry: %w", err)
	}

	return ox, oy, int(analysis.GetImageWidth()), int(analysis.GetImageHeight()), nil
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

// ShrinkMaintainAspectRatio shrinks an input wand to fit within a box of given width and height if it exceeds those dimensions, maintaining aspect ratio
func ShrinkMaintainAspectRatio(wand *imagick.MagickWand, width uint, height uint) error {
	inputHeight := float64(wand.GetImageHeight())
	inputWidth := float64(wand.GetImageWidth())

	if inputWidth <= float64(width) && inputHeight <= float64(height) {
		return nil
	}

	return ResizeMaintainAspectRatio(wand, width, height)
}

type OverlayOptions struct {
	HFlip bool
	VFlip bool

	OverlayWidthFactor  float64
	OverlayHeightFactor float64

	RightToLeft bool
}

// FrameArgs are the arguments for frame commands.
type FrameArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args FrameArgs) GetImageURL() string {
	return args.ImageURL
}

// FitMode controls how the input image is resized to fill the frame opening.
type FitMode int

const (
	// FitModeFit resizes to fit within the opening, maintaining aspect ratio.
	FitModeFit FitMode = iota
	// FitModeStretch stretches the image to exactly fill the opening, ignoring aspect ratio.
	FitModeStretch
	// FitModeFitHeight resizes so the height matches the opening, maintaining aspect ratio.
	// The width may exceed the opening width and will be clipped.
	FitModeFitHeight
)

// PositionMode controls where the input image is placed within the frame opening.
type PositionMode int

const (
	// PositionModeCentered centers the image within the opening.
	PositionModeCentered PositionMode = iota
	// PositionModeTopLeft places the image in the top-left corner of the opening.
	PositionModeTopLeft
)

// FrameOptions controls the behaviour of a frame command.
type FrameOptions struct {
	FitMode      FitMode
	PositionMode PositionMode
}

// FrameImage places wand into the given opening of a frame, resizing according to options,
// centering on a white background, and compositing behind the frame.
func FrameImage(wand *imagick.MagickWand, frame *imagick.MagickWand, openX, openY, openW, openH int, options FrameOptions) ([]*imagick.MagickWand, error) {
	switch options.FitMode {
	case FitModeStretch:
		if err := wand.ResizeImage(uint(openW), uint(openH), imagick.FILTER_LANCZOS); err != nil {
			return nil, fmt.Errorf("error resizing image: %w", err)
		}
	case FitModeFit:
		if err := ResizeMaintainAspectRatio(wand, uint(openW), uint(openH)); err != nil {
			return nil, fmt.Errorf("error resizing image: %w", err)
		}
	case FitModeFitHeight:
		scale := float64(openH) / float64(wand.GetImageHeight())
		newW := uint(float64(wand.GetImageWidth()) * scale)
		if err := wand.ResizeImage(newW, uint(openH), imagick.FILTER_LANCZOS); err != nil {
			return nil, fmt.Errorf("error resizing image: %w", err)
		}
	}

	bg := imagick.NewMagickWand()
	bgColor := imagick.NewPixelWand()
	bgColor.SetColor("white")
	if err := bg.NewImage(uint(openW), uint(openH), bgColor); err != nil {
		return nil, fmt.Errorf("error creating background: %w", err)
	}

	x, y := 0, 0
	if options.PositionMode == PositionModeCentered {
		x = (openW - int(wand.GetImageWidth())) / 2
		y = (openH - int(wand.GetImageHeight())) / 2
	}
	if err := bg.CompositeImage(wand, imagick.COMPOSITE_OP_OVER, true, x, y); err != nil {
		return nil, fmt.Errorf("error compositing image onto background: %w", err)
	}

	if err := frame.CompositeImage(bg, imagick.COMPOSITE_OP_DST_OVER, true, openX, openY); err != nil {
		return nil, fmt.Errorf("error compositing background onto frame: %w", err)
	}

	return []*imagick.MagickWand{frame}, nil
}

// MakeImageFrameOp creates an ImageOperation that places the input image into the transparent
// opening of a frame image, auto-detecting the opening position and size.
func MakeImageFrameOp(frameBytes []byte, options FrameOptions) ImageOperation[FrameArgs] {
	return func(wand *imagick.MagickWand, args FrameArgs) ([]*imagick.MagickWand, error) {
		frame := imagick.NewMagickWand()
		if err := frame.ReadImageBlob(frameBytes); err != nil {
			return nil, fmt.Errorf("error reading frame: %w", err)
		}

		openX, openY, openW, openH, err := FindTransparentOpeningRect(frame)
		if err != nil {
			return nil, fmt.Errorf("error finding frame opening: %w", err)
		}

		return FrameImage(wand, frame, openX, openY, openW, openH, options)
	}
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

func MakeImageOverlayOp(overlayImage []byte, initialOptions OverlayOptions) ImageOperation[OverlayImageArgs] {
	return func(wand *imagick.MagickWand, args OverlayImageArgs) ([]*imagick.MagickWand, error) {
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
	}
}

