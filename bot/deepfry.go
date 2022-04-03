package bot

import (
	"fmt"
	"io"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type DeepfryArgs struct {
	ImageURL        string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius      float64 `default:"100" description:"Radius of outline to draw around edges."`
	DownscaleFactor uint    `default:"2" description:"Factor to downscale the image by while processing."`
}

func (args DeepfryArgs) GetImageURL() string {
	return args.ImageURL
}

// Deepfry destroys an image via a combination of operations.
func Deepfry(src []byte, dest io.Writer, args DeepfryArgs) error {
	wand := imagick.NewMagickWand()
	err := wand.ReadImageBlob(src)
	if err != nil {
		return fmt.Errorf("error reading image: %w", err)
	}

	err = wand.ResizeImage(wand.GetImageWidth()/args.DownscaleFactor, wand.GetImageHeight()/args.DownscaleFactor, imagick.FILTER_CUBIC, 0.5)
	if err != nil {
		return fmt.Errorf("error resizing image: %w", err)
	}

	err = wand.ResizeImage(wand.GetImageWidth()*args.DownscaleFactor, wand.GetImageHeight()*args.DownscaleFactor, imagick.FILTER_CUBIC, 0.5)
	if err != nil {
		return fmt.Errorf("error resizing image: %w", err)
	}

	err = wand.EdgeImage(args.EdgeRadius)
	if err != nil {
		return fmt.Errorf("error edge enhancing image: %w", err)
	}

	_, err = dest.Write(wand.GetImageBlob())
	if err != nil {
		return fmt.Errorf("error writing image: %w", err)
	}

	return nil
}
