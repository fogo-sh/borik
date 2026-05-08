package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Deepfry(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var deepfryArgs args.Deepfry
	err = decodeOperationArgs(opArgs, &deepfryArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	err = wand.ResizeImage(wand.GetImageWidth()/deepfryArgs.DownscaleFactor, wand.GetImageHeight()/deepfryArgs.DownscaleFactor, imagick.FILTER_POINT)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	err = wand.ResizeImage(wand.GetImageWidth()*deepfryArgs.DownscaleFactor, wand.GetImageHeight()*deepfryArgs.DownscaleFactor, imagick.FILTER_POINT)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	wand.SetImageChannelMask(imagick.CHANNEL_RED | imagick.CHANNEL_GREEN | imagick.CHANNEL_BLUE)

	err = wand.EdgeImage(deepfryArgs.EdgeRadius)
	if err != nil {
		return nil, fmt.Errorf("error edge enhancing image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
