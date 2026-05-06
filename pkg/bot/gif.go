package bot

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type GifArgs struct {
	VideoURL string  `default:"" description:"URL to the video to process. Leave blank to automatically attempt to find a video."`
	FPS      uint    `default:"15" description:"Frames per second for the GIF."`
	Width    uint    `default:"480" description:"Width of the GIF in pixels. Height is scaled to preserve aspect ratio."`
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

	srcBytes, err := DownloadImage(videoURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download video to process")
		return
	}

	parsedURL, _ := url.Parse(videoURL)
	inputFile, err := os.CreateTemp("", "borik-gif-input-*"+path.Ext(parsedURL.Path))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temporary video file")
		return
	}
	inputPath := inputFile.Name()
	defer func() {
		if err := os.Remove(inputPath); err != nil {
			log.Error().Err(err).Msg("Failed to remove temporary video file")
		}
	}()

	if _, err := inputFile.Write(srcBytes); err != nil {
		log.Error().Err(err).Msg("Failed to write temporary video file")
		_ = inputFile.Close()
		return
	}
	if err := inputFile.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close temporary video file")
		return
	}

	outputFile, err := os.CreateTemp("", "borik-gif-output-*.gif")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temporary GIF file")
		return
	}
	outputPath := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close temporary GIF file")
		return
	}
	defer func() {
		if err := os.Remove(outputPath); err != nil {
			log.Error().Err(err).Msg("Failed to remove temporary GIF file")
		}
	}()

	if err := convertVideoToGIF(inputPath, outputPath, args); err != nil {
		log.Error().Err(err).Msg("Failed to convert video to GIF")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to convert video to GIF: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
		return
	}

	gifBytes, err := os.ReadFile(outputPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read output GIF")
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
			Reader:      bytes.NewReader(gifBytes),
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send GIF")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to send resulting GIF: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
	}
}

func convertVideoToGIF(inputPath string, outputPath string, args GifArgs) error {
	if args.FPS == 0 {
		return fmt.Errorf("fps must be greater than 0")
	}
	if args.Width == 0 {
		return fmt.Errorf("width must be greater than 0")
	}

	ffmpegArgs := []string{
		"-hide_banner",
		"-loglevel", "error",
	}
	if args.Duration > 0 {
		ffmpegArgs = append(ffmpegArgs, "-t", fmt.Sprintf("%.3f", args.Duration))
	}

	filter := fmt.Sprintf(
		"fps=%d,scale=%d:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse",
		args.FPS,
		args.Width,
	)
	ffmpegArgs = append(ffmpegArgs,
		"-i", inputPath,
		"-vf", filter,
		"-loop", "0",
		"-y",
		outputPath,
	)

	output, err := exec.Command("ffmpeg", ffmpegArgs...).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("ffmpeg failed: %s", message)
	}

	return nil
}
