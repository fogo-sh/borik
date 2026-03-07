package bot

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

// RendererOperation defines an animation rendered by the node-renderer sidecar.
type RendererOperation struct {
	Animation string
	Params    func(args RendererArgs) map[string]any
}

type RendererArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Frames   uint   `default:"24" description:"Number of frames in the output GIF."`
	Size     uint   `default:"400" description:"Canvas size in pixels."`
}

func (args RendererArgs) GetImageURL() string {
	return args.ImageURL
}

func invokeRendererOperation(ctx *OperationContext, args RendererArgs, op RendererOperation) {
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

	params := map[string]any{
		"totalFrames": args.Frames,
		"size":        args.Size,
	}
	if op.Params != nil {
		for k, v := range op.Params(args) {
			params[k] = v
		}
	}

	frames, delay, err := RenderAnimation(op.Animation, srcBytes, params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render animation")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to render animation: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error response")
		}
		return
	}

	resultImage := imagick.NewMagickWand()
	for i, frameBytes := range frames {
		frameWand := imagick.NewMagickWand()
		if err := frameWand.ReadImageBlob(frameBytes); err != nil {
			log.Error().Err(err).Int("frame", i).Msg("Failed to read rendered frame")
			return
		}
		if err := resultImage.AddImage(frameWand); err != nil {
			log.Error().Err(err).Int("frame", i).Msg("Failed to add frame to result")
			return
		}
	}

	resultImage.ResetIterator()

	if err := resultImage.SetImageFormat("GIF"); err != nil {
		log.Error().Err(err).Msg("Failed to set result format")
		return
	}

	// delay from renderer is in ms; ImageMagick uses centiseconds
	if err := resultImage.SetImageDelay(uint(delay / 10)); err != nil {
		log.Error().Err(err).Msg("Failed to set frame delay")
		return
	}

	if err := resultImage.ResetImagePage("0x0+0+0"); err != nil {
		log.Error().Err(err).Msg("Failed to repage image")
	}

	resultImage = resultImage.DeconstructImages()

	imageBlob, err := resultImage.GetImagesBlob()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image blob")
		return
	}

	destBuffer := new(bytes.Buffer)
	if _, err := destBuffer.Write(imageBlob); err != nil {
		log.Error().Err(err).Msg("Failed to write image blob")
		return
	}

	resultFileName := fmt.Sprintf("%s.gif", strings.ReplaceAll(op.Animation, "-", "_"))
	if err := ctx.SendFiles([]*discordgo.File{{
		Name:   resultFileName,
		Reader: destBuffer,
	}}); err != nil {
		log.Error().Err(err).Msg("Failed to send rendered GIF")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to send resulting image: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
	}
}

func MakeRendererTextCommand(op RendererOperation) func(*discordgo.MessageCreate, RendererArgs) {
	return func(message *discordgo.MessageCreate, args RendererArgs) {
		invokeRendererOperation(NewOperationContextFromMessage(Instance.session, message), args, op)
	}
}

func MakeRendererSlashCommand(op RendererOperation) func(*discordgo.Session, *discordgo.InteractionCreate, RendererArgs) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, args RendererArgs) {
		invokeRendererOperation(NewOperationContextFromInteraction(session, interaction), args, op)
	}
}

var (
	SphereOp        = RendererOperation{Animation: "rotating-sphere"}
	InsideSphereOp  = RendererOperation{Animation: "inside-sphere"}
	LowPolySphereOp = RendererOperation{Animation: "low-poly-sphere"}
	PyramidOp       = RendererOperation{Animation: "pyramid"}
	Spin360Op       = RendererOperation{Animation: "360-spin"}
)
