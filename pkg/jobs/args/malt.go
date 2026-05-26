package args

type Malt struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degree   float64 `default:"45" description:"Number of degrees to rotate the image by while processing."`
}

func (a Malt) GetImageURL() string {
	return a.ImageURL
}

func (a Malt) ActivityName() string {
	return "Malt"
}
