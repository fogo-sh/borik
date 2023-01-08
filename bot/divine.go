package bot

import (
	_ "embed"
	"fmt"

	imagick7 "gopkg.in/gographics/imagick.v3/imagick"
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

func Divine(wand *imagick7.MagickWand, args DivineArgs) ([]*imagick7.MagickWand, error) {
	overlay := imagick7.NewMagickWand()
	err := overlay.ReadImageBlob(divineOverlayImage)
	if err != nil {
		return nil, fmt.Errorf("error reading divine overlay image: %w", err)
	}

	wand.SetImageChannelMask(imagick7.CHANNEL_BLUE | imagick7.CHANNEL_GREEN)
	err = wand.EvaluateImage(imagick7.EVAL_OP_SET, 0)
	if err != nil {
		return nil, fmt.Errorf("error removing blue & green channels: %w", err)
	}

	wand.SetImageChannelMask(imagick7.CHANNEL_RED | imagick7.CHANNEL_GREEN | imagick7.CHANNEL_BLUE)

	// TODO: Figure out why this is making everything white
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

	err = ResizeMaintainAspectRatio(overlay, inputWidth, inputHeight)
	if err != nil {
		return nil, fmt.Errorf("error resizing overlay: %w", err)
	}

	overlayWidth := overlay.GetImageWidth()
	overlayHeight := overlay.GetImageHeight()

	err = wand.CompositeImage(
		overlay,
		imagick7.COMPOSITE_OP_ATOP,
		false,
		int((inputWidth/2)-(overlayWidth/2)),
		int((inputHeight/2)-(overlayHeight/2)),
	)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick7.MagickWand{wand}, nil
}
