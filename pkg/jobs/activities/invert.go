package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Invert(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	wand.SetImageChannelMask(imagick.CHANNEL_RED | imagick.CHANNEL_GREEN | imagick.CHANNEL_BLUE)
	err = wand.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error inverting image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
