package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Magik(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var magikArgs args.Magik
	err = decodeOperationArgs(opArgs, &magikArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	return magikHelper(jobWorkspace, wand, magikArgs.Scale, magikArgs.WidthMultiplier, magikArgs.HeightMultiplier)
}

func Lagik(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var lagikArgs args.Lagik
	err = decodeOperationArgs(opArgs, &lagikArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	return magikHelper(jobWorkspace, wand, lagikArgs.Scale, 1.5, 1.5)
}

func Gmagik(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var gmagikArgs args.Gmagik
	err = decodeOperationArgs(opArgs, &gmagikArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	var results []workspace.Artifact
	lastFrame := wand

	for i := uint(0); i < gmagikArgs.Iterations; i++ {
		newFrame := lastFrame.Clone()

		frames, err := magikHelper(jobWorkspace, newFrame, gmagikArgs.Scale, gmagikArgs.WidthMultiplier, gmagikArgs.HeightMultiplier)
		if err != nil {
			return nil, fmt.Errorf("error running magik: %w", err)
		}

		lastFrame = newFrame
		results = append(results, frames...)
	}

	return results, nil
}

func magikHelper(jobWorkspace workspace.Workspace, wand *imagick.MagickWand, scale, widthMultiplier, heightMultiplier float64) ([]workspace.Artifact, error) {
	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err := wand.LiquidRescaleImage(uint(float64(width)*widthMultiplier), uint(float64(height)*heightMultiplier), scale, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	err = wand.ResizeImage(width, height, imagick.FILTER_LANCZOS)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to resize image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
