package bot

import (
	"gopkg.in/gographics/imagick.v2/imagick"
)

type TransformArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Crop     string `default:"" description:"Crop geometry to apply to the image. Leave blank to not crop."`
	Size     string `default:"" description:"Size geometry to apply to the image. Leave blank to not resize."`
}

func (args TransformArgs) GetImageURL() string {
	return args.ImageURL
}

func Transform(wand *imagick.MagickWand, args TransformArgs) ([]*imagick.MagickWand, error) {
	transformedImage := wand.TransformImage(args.Crop, args.Size)

	return []*imagick.MagickWand{transformedImage}, nil
}
