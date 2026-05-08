package bot

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	jobArgs "github.com/fogo-sh/borik/pkg/jobs/args"
)

type GifArgs struct {
	VideoURL string  `default:"" description:"URL to the video to process. Leave blank to automatically attempt to find a video."`
	FPS      uint    `default:"10" description:"Frames per second for the GIF."`
	Width    uint    `default:"320" description:"Width of the GIF in pixels. Height is scaled to preserve aspect ratio."`
	Duration float64 `default:"10" description:"Maximum video duration to convert, in seconds. Set to 0 to convert the whole video."`
}

// GifTextCommand converts a video to a GIF from a text command.
func GifTextCommand(message *discordgo.MessageCreate, args GifArgs) {
	PrepareAndInvokeGif(NewOperationContextFromMessage(Instance.session, message), args)
}

// GifSlashCommand converts a video to a GIF from a slash command.
func GifSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args GifArgs) {
	PrepareAndInvokeGif(NewOperationContextFromInteraction(session, interaction), args)
}

// PrepareAndInvokeGif locates a video, converts it to a GIF, and uploads the result.
func PrepareAndInvokeGif(ctx *OperationContext, args GifArgs) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	videoURL := args.VideoURL
	if videoURL == "" {
		var err error
		videoURL, err = ctx.FindVideoURL()
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find video to process")
			return
		}
	}

	parsedURL, _ := url.Parse(videoURL)

	_, resultReader, err := Instance.triggerGif(
		context.Background(),
		ctx.GetSourceID(),
		jobArgs.Gif{
			VideoURL: videoURL,
			FPS:      args.FPS,
			Width:    args.Width,
			Duration: args.Duration,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert video to GIF")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to convert video to GIF: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
		return
	}

	originalFileName := path.Base(parsedURL.Path)
	if originalFileName == "." || originalFileName == "/" {
		originalFileName = "video"
	}
	originalFileNameNoExt := strings.TrimSuffix(originalFileName, path.Ext(originalFileName))
	resultFileName := fmt.Sprintf("%s.gif", originalFileNameNoExt)

	log.Debug().Msg("GIF processed, uploading result")
	err = ctx.SendFiles([]*discordgo.File{
		{
			Name:        resultFileName,
			ContentType: "image/gif",
			Reader:      resultReader,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send GIF")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to send resulting GIF: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
	}
}
