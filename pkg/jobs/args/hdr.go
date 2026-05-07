package args

type Hdr struct {
	ImageURL      string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Multiply      float64 `default:"1.5" description:"Multiplier for pixel values. Higher values produce brighter, more saturated results."`
	GammaExponent float64 `default:"0.9" description:"Exponent for gamma power curve. Lower values brighten midtones more."`
}

func (a Hdr) GetImageURL() string {
	return a.ImageURL
}

func (a Hdr) ActivityName() string {
	return "Hdr"
}
