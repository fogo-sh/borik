package args

type Deepfry struct {
	ImageURL        string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius      float64 `default:"100" description:"Radius of outline to draw around edges."`
	DownscaleFactor uint    `default:"2" description:"Factor to downscale the image by while processing."`
}

func (a Deepfry) GetImageURL() string {
	return a.ImageURL
}

func (a Deepfry) ActivityName() string {
	return "Deepfry"
}
