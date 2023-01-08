package bot

import (
	imagick7 "gopkg.in/gographics/imagick.v3/imagick"
)

type V7TestArgs struct {
	ImageURL string `default:""`
}

func (a V7TestArgs) GetImageURL() string {
	return a.ImageURL
}

func V7Test(wand *imagick7.MagickWand, _ V7TestArgs) ([]*imagick7.MagickWand, error) {
	return []*imagick7.MagickWand{wand}, nil
}
