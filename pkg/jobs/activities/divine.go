package activities

import (
	"context"
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

//go:embed images/overlays/divine.png
var divineOverlayImage []byte

func Divine(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var divineArgs args.Divine
	err = decodeOperationArgs(opArgs, &divineArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	overlay := imagick.NewMagickWand()
	defer overlay.Destroy()

	err = overlay.ReadImageBlob(divineOverlayImage)
	if err != nil {
		return nil, fmt.Errorf("error reading divine overlay image: %w", err)
	}

	err = wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_OPAQUE)
	if err != nil {
		return nil, fmt.Errorf("error removing alpha: %w", err)
	}

	wand.SetImageChannelMask(imagick.CHANNEL_BLUE | imagick.CHANNEL_GREEN)
	err = wand.EvaluateImage(imagick.EVAL_OP_SET, 0)
	if err != nil {
		return nil, fmt.Errorf("error removing blue & green channels: %w", err)
	}

	wand.SetImageChannelMask(imagick.CHANNEL_RED | imagick.CHANNEL_GREEN | imagick.CHANNEL_BLUE)
	err = wand.EdgeImage(divineArgs.EdgeRadius)
	if err != nil {
		return nil, fmt.Errorf("error edge detecting: %w", err)
	}

	wand.SetImageChannelMask(imagick.CHANNELS_DEFAULT)

	err = wand.ClampImage()
	if err != nil {
		return nil, fmt.Errorf("error clamping image: %w", err)
	}

	err = wand.ModulateImage(divineArgs.Brightness, divineArgs.Saturation, divineArgs.Hue)
	if err != nil {
		return nil, fmt.Errorf("error decreasing saturation: %w", err)
	}

	err = wand.GaussianBlurImage(divineArgs.BlurRadius, divineArgs.BlurSigma)
	if err != nil {
		return nil, fmt.Errorf("error blurring image: %w", err)
	}

	inputHeight := wand.GetImageHeight()
	inputWidth := wand.GetImageWidth()

	err = resizeMaintainAspectRatio(overlay, inputWidth, inputHeight)
	if err != nil {
		return nil, fmt.Errorf("error resizing overlay: %w", err)
	}

	overlayWidth := overlay.GetImageWidth()
	overlayHeight := overlay.GetImageHeight()

	err = wand.CompositeImage(
		overlay,
		imagick.COMPOSITE_OP_ATOP,
		true,
		int((inputWidth/2)-(overlayWidth/2)),
		int((inputHeight/2)-(overlayHeight/2)),
	)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
