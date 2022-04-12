package bot

import (
	"context"
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v2/imagick"
)

//go:embed divine.png
var divineOverlayImage []byte

type DivineArgs struct {
	ImageURL   string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius float64 `default:"5" description:"Edge radius for edge detection."`
	BlurRadius float64 `default:"4" description:"Gaussian blur radius."`
	BlurSigma  float64 `default:"2" description:"Sigma value for gaussian blur."`
	Brightness float64 `default:"100" description:"Relative percentage for the brightness of the final image."`
	Saturation float64 `default:"50" description:"Relative percentage for the saturation of the final image."`
	Hue        float64 `default:"100" description:"Relative percentage for the hue of the final image."`
}

func (args DivineArgs) GetImageURL() string {
	return args.ImageURL
}

func Divine(ctx context.Context, wand *imagick.MagickWand, args DivineArgs) ([]*imagick.MagickWand, error) {
	overlay := imagick.NewMagickWand()
	err := overlay.ReadImageBlob(divineOverlayImage)
	if err != nil {
		return nil, fmt.Errorf("error reading divine overlay image: %w", err)
	}

	err = wand.EvaluateImageChannel(imagick.CHANNEL_BLUE, imagick.EVAL_OP_SET, 0)
	if err != nil {
		return nil, fmt.Errorf("error removing blue channel: %w", err)
	}

	err = wand.EvaluateImageChannel(imagick.CHANNEL_GREEN, imagick.EVAL_OP_SET, 0)
	if err != nil {
		return nil, fmt.Errorf("error removing green channel: %w", err)
	}

	err = wand.EdgeImage(args.EdgeRadius)
	if err != nil {
		return nil, fmt.Errorf("error edge detecting: %w", err)
	}

	err = wand.ModulateImage(args.Brightness, args.Saturation, args.Hue)
	if err != nil {
		return nil, fmt.Errorf("error decreasing saturation: %w", err)
	}

	err = wand.GaussianBlurImage(args.BlurRadius, args.BlurSigma)
	if err != nil {
		return nil, fmt.Errorf("error blurring image: %w", err)
	}

	inputHeight := wand.GetImageHeight()
	inputWidth := wand.GetImageWidth()

	overlay = overlay.TransformImage("", fmt.Sprintf("%dx%d", inputWidth, inputHeight))

	overlayWidth := overlay.GetImageWidth()
	overlayHeight := overlay.GetImageHeight()

	err = wand.CompositeImage(
		overlay,
		imagick.COMPOSITE_OP_ATOP,
		int((inputWidth/2)-(overlayWidth/2)),
		int((inputHeight/2)-(overlayHeight/2)),
	)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
