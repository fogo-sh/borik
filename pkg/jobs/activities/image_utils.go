package activities

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func LoadImage(ctx context.Context, jobWorkspace workspace.Workspace, imageUrl string) (workspace.Artifact, error) {
	resp, err := http.Get(imageUrl)
	if err != nil {
		return "", fmt.Errorf("error downloading image: %w", err)
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)

	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error copying image to buffer: %w", err)
	}

	artifact, err := jobWorkspace.Persist(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("error persisting image: %w", err)
	}

	return artifact, nil
}

func SplitImage(ctx context.Context, jobWorkspace workspace.Workspace, inputArtifact workspace.Artifact) ([]workspace.Artifact, error) {
	input, err := jobWorkspace.RetrieveWand(inputArtifact)
	if err != nil {
		return nil, fmt.Errorf("error retrieving image: %w", err)
	}
	defer input.Destroy()

	var resultArtifacts []workspace.Artifact

	for i := uint(0); i < input.GetNumberImages(); i++ {
		input.SetIteratorIndex(int(i))
		artifact, err := jobWorkspace.Persist(input.GetImageBlob())
		if err != nil {
			return nil, fmt.Errorf("error persisting frame: %w", err)
		}

		resultArtifacts = append(resultArtifacts, artifact)
	}

	return resultArtifacts, nil
}

func JoinImage(ctx context.Context, jobWorkspace workspace.Workspace, inputArtifacts []workspace.Artifact) (workspace.Artifact, error) {
	output := imagick.NewMagickWand()
	defer output.Destroy()

	for _, artifact := range inputArtifacts {
		frame, err := jobWorkspace.RetrieveWand(artifact)
		if err != nil {
			return "", fmt.Errorf("error retrieving frame: %w", err)
		}

		err = output.AddImage(frame)
		if err != nil {
			return "", fmt.Errorf("error adding frame to output: %w", err)
		}
	}

	output.ResetIterator()

	if len(inputArtifacts) > 1 {
		err := output.SetImageFormat("GIF")
		if err != nil {
			return "", fmt.Errorf("error setting output format: %w", err)
		}

		err = output.SetImageDelay(10)
		if err != nil {
			return "", fmt.Errorf("error setting output framerate: %w", err)
		}
	} else {
		err := output.SetImageFormat("PNG")
		if err != nil {
			return "", fmt.Errorf("error setting output format: %w", err)
		}
	}

	err := output.ResetImagePage("0x0+0+0")
	if err != nil {
		return "", fmt.Errorf("error repaging output: %w", err)
	}

	output = output.DeconstructImages()

	outputArtifact, err := jobWorkspace.Persist(output.GetImagesBlob())
	if err != nil {
		return "", fmt.Errorf("error persisting output: %w", err)
	}

	return outputArtifact, nil
}

func loadFrame(frameBytes []byte) (*imagick.MagickWand, error) {
	frame := imagick.NewMagickWand()
	err := frame.ReadImageBlob(frameBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading frame: %w", err)
	}

	return frame, nil
}

func saveFrames(jobWorkspace workspace.Workspace, frames ...*imagick.MagickWand) ([]workspace.Artifact, error) {
	var resultArtifacts []workspace.Artifact

	for _, frame := range frames {
		artifact, err := jobWorkspace.PersistWand(frame)
		if err != nil {
			return nil, fmt.Errorf("error persisting frame: %w", err)
		}
		resultArtifacts = append(resultArtifacts, artifact)
	}

	return resultArtifacts, nil
}
