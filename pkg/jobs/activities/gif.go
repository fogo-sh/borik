package activities

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func ConvertVideoToGIF(ctx context.Context, jobWorkspace workspace.Workspace, args args.Gif) (workspace.Artifact, error) {
	inputPath, err := downloadVideo(ctx, jobWorkspace, args.VideoURL)
	if err != nil {
		return "", err
	}
	defer removeTempFile(inputPath)

	outputFile, err := os.CreateTemp(jobWorkspace.Path, "borik-gif-output-*.gif")
	if err != nil {
		return "", fmt.Errorf("error creating temporary GIF file: %w", err)
	}
	outputPath := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		return "", fmt.Errorf("error closing temporary GIF file: %w", err)
	}
	defer removeTempFile(outputPath)

	if err := convertVideoToGIF(ctx, inputPath, outputPath, args); err != nil {
		return "", err
	}

	gifBytes, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("error reading output GIF: %w", err)
	}

	outputArtifact, err := jobWorkspace.Persist(gifBytes)
	if err != nil {
		return "", fmt.Errorf("error persisting output GIF: %w", err)
	}

	return outputArtifact, nil
}

func downloadVideo(ctx context.Context, jobWorkspace workspace.Workspace, videoURL string) (string, error) {
	parsedURL, err := url.Parse(videoURL)
	if err != nil {
		return "", fmt.Errorf("error parsing video URL: %w", err)
	}

	inputFile, err := os.CreateTemp(jobWorkspace.Path, "borik-gif-input-*"+path.Ext(parsedURL.Path))
	if err != nil {
		return "", fmt.Errorf("error creating temporary video file: %w", err)
	}
	inputPath := inputFile.Name()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, videoURL, nil)
	if err != nil {
		_ = inputFile.Close()
		removeTempFile(inputPath)
		return "", fmt.Errorf("error creating video request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = inputFile.Close()
		removeTempFile(inputPath)
		return "", fmt.Errorf("error downloading video: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_ = inputFile.Close()
		removeTempFile(inputPath)
		return "", fmt.Errorf("error downloading video: unexpected status %s", resp.Status)
	}

	_, err = io.Copy(inputFile, resp.Body)
	if err != nil {
		_ = inputFile.Close()
		removeTempFile(inputPath)
		return "", fmt.Errorf("error writing temporary video file: %w", err)
	}

	if err := inputFile.Close(); err != nil {
		removeTempFile(inputPath)
		return "", fmt.Errorf("error closing temporary video file: %w", err)
	}

	return inputPath, nil
}

func convertVideoToGIF(ctx context.Context, inputPath string, outputPath string, args args.Gif) error {
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

	output, err := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("ffmpeg failed: %s", message)
	}

	return nil
}

func removeTempFile(path string) {
	_ = os.Remove(path)
}
