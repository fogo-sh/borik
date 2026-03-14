package bot

import (
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed heritage.png
var heritageImage []byte

type HeritageArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args HeritageArgs) GetImageURL() string {
	return args.ImageURL
}

func Heritage(wand *imagick.MagickWand, args HeritageArgs) ([]*imagick.MagickWand, error) {
	frame := imagick.NewMagickWand()
	if err := frame.ReadImageBlob(heritageImage); err != nil {
		return nil, fmt.Errorf("error reading frame: %w", err)
	}

	openX, openY, openW, openH, err := FindTransparentOpeningRect(frame)
	if err != nil {
		return nil, fmt.Errorf("error finding frame opening: %w", err)
	}

	if err := ResizeMaintainAspectRatio(wand, uint(openW), uint(openH)); err != nil {
		return nil, fmt.Errorf("error resizing original image: %w", err)
	}

	bg := imagick.NewMagickWand()
	bgColor := imagick.NewPixelWand()
	bgColor.SetColor("white")
	if err := bg.NewImage(uint(openW), uint(openH), bgColor); err != nil {
		return nil, fmt.Errorf("error creating background: %w", err)
	}

	x := (openW - int(wand.GetImageWidth())) / 2
	y := (openH - int(wand.GetImageHeight())) / 2
	if err := bg.CompositeImage(wand, imagick.COMPOSITE_OP_OVER, true, x, y); err != nil {
		return nil, fmt.Errorf("error compositing wand onto background: %w", err)
	}

	if err := frame.CompositeImage(bg, imagick.COMPOSITE_OP_DST_OVER, true, openX, openY); err != nil {
		return nil, fmt.Errorf("error compositing background onto frame: %w", err)
	}

	return []*imagick.MagickWand{frame}, nil
}
