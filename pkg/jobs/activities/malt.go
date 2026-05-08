package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Malt(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var maltArgs args.Malt
	err = decodeOperationArgs(opArgs, &maltArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err = wand.SwirlImage(maltArgs.Degree, imagick.INTERPOLATE_PIXEL_BILINEAR)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to swirl: %w", err)
	}

	err = wand.LiquidRescaleImage(width/2, height/2, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	err = wand.SwirlImage(maltArgs.Degree*-1, imagick.INTERPOLATE_PIXEL_BILINEAR)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to swirl: %w", err)
	}

	err = wand.LiquidRescaleImage(width, height, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
