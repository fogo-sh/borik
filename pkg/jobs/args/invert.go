package args

type Invert struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Invert) GetImageURL() string {
	return a.ImageURL
}

func (a Invert) ActivityName() string {
	return "Invert"
}
