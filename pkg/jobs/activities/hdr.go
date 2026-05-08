package activities

import (
	"context"
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

//go:embed profiles/2020_profile.icc
var icc2020Profile []byte

func Hdr(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var hdrArgs args.Hdr
	err = decodeOperationArgs(opArgs, &hdrArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	err = wand.TransformImageColorspace(imagick.COLORSPACE_RGB)
	if err != nil {
		return nil, fmt.Errorf("error converting to RGB colorspace: %w", err)
	}

	err = wand.AutoGammaImage()
	if err != nil {
		return nil, fmt.Errorf("error applying auto-gamma: %w", err)
	}

	err = wand.EvaluateImage(imagick.EVAL_OP_MULTIPLY, hdrArgs.Multiply)
	if err != nil {
		return nil, fmt.Errorf("error evaluating multiply: %w", err)
	}

	err = wand.EvaluateImage(imagick.EVAL_OP_POW, hdrArgs.GammaExponent)
	if err != nil {
		return nil, fmt.Errorf("error evaluating pow: %w", err)
	}

	err = wand.TransformImageColorspace(imagick.COLORSPACE_SRGB)
	if err != nil {
		return nil, fmt.Errorf("error converting to sRGB colorspace: %w", err)
	}

	err = wand.SetImageDepth(16)
	if err != nil {
		return nil, fmt.Errorf("error setting image depth: %w", err)
	}

	err = wand.ProfileImage("icc", icc2020Profile)
	if err != nil {
		return nil, fmt.Errorf("error applying ICC profile: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
