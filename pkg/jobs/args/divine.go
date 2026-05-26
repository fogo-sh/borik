package args

type Divine struct {
	ImageURL   string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius float64 `default:"5" description:"Edge radius for edge detection."`
	BlurRadius float64 `default:"4" description:"Gaussian blur radius."`
	BlurSigma  float64 `default:"2" description:"Sigma value for gaussian blur."`
	Brightness float64 `default:"100" description:"Relative percentage for the brightness of the final image."`
	Saturation float64 `default:"50" description:"Relative percentage for the saturation of the final image."`
	Hue        float64 `default:"100" description:"Relative percentage for the hue of the final image."`
}

func (a Divine) GetImageURL() string {
	return a.ImageURL
}

func (a Divine) ActivityName() string {
	return "Divine"
}
