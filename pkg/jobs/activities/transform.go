package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Rotate(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var rotateArgs args.Rotate
	err = decodeOperationArgs(opArgs, &rotateArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	bgWand := imagick.NewPixelWand()
	defer bgWand.Destroy()
	bgWand.SetAlpha(0)

	err = wand.RotateImage(bgWand, rotateArgs.Degrees)
	if err != nil {
		return nil, fmt.Errorf("error rotating image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}

func Resize(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var resizeArgs args.Resize
	err = decodeOperationArgs(opArgs, &resizeArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	var targetHeight, targetWidth uint
	switch resizeArgs.Mode {
	case "absolute":
		targetHeight = uint(resizeArgs.Height)
		targetWidth = uint(resizeArgs.Width)
	case "percent":
		targetHeight = uint((resizeArgs.Height / 100) * float64(wand.GetImageHeight()))
		targetWidth = uint((resizeArgs.Width / 100) * float64(wand.GetImageWidth()))
	default:
		return nil, fmt.Errorf("unsupported mode: %s (must be one of percent, absolute)", resizeArgs.Mode)
	}

	err = wand.ResizeImage(targetWidth, targetHeight, imagick.FILTER_POINT)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
