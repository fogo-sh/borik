package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func APNGToGIF(ctx context.Context, jobWorkspace workspace.Workspace, args args.APNGToGIF) (workspace.Artifact, error) {
	inputArtifact, err := LoadImage(ctx, jobWorkspace, args.ImageURL)
	if err != nil {
		return "", err
	}

	input, err := jobWorkspace.Retrieve(inputArtifact)
	if err != nil {
		return "", fmt.Errorf("error retrieving APNG: %w", err)
	}

	wand := imagick.NewMagickWand()
	defer func() {
		wand.Destroy()
	}()

	err = wand.SetFilename("APNG:profile.png")
	if err != nil {
		return "", fmt.Errorf("error setting input format: %w", err)
	}

	err = wand.ReadImageBlob(input)
	if err != nil {
		return "", fmt.Errorf("error reading input image: %w", err)
	}

	for i := uint(0); i < wand.GetNumberImages(); i++ {
		wand.SetIteratorIndex(int(i))
		err = wand.SetImageDispose(imagick.DISPOSE_BACKGROUND)
		if err != nil {
			return "", fmt.Errorf("error configuring disposal: %w", err)
		}
	}

	wand.ResetIterator()
	coalesced := wand.CoalesceImages()
	oldWand := wand
	wand = coalesced
	oldWand.Destroy()

	err = wand.SetFilename("profile.gif")
	if err != nil {
		return "", fmt.Errorf("error setting output format: %w", err)
	}

	imageBlob, err := wand.GetImagesBlob()
	if err != nil {
		return "", fmt.Errorf("error generating output image: %w", err)
	}

	if len(imageBlob) == 0 {
		return "", fmt.Errorf("got an empty output image - your provided sticker may be one of the currently broken ones")
	}

	outputArtifact, err := jobWorkspace.Persist(imageBlob)
	if err != nil {
		return "", fmt.Errorf("error persisting output GIF: %w", err)
	}

	return outputArtifact, nil
}
