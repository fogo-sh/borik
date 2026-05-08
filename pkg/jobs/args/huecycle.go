package args

type HueCycle struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Steps    uint   `default:"20" description:"Number of steps to do the hue shift in."`
}

func (a HueCycle) GetImageURL() string {
	return a.ImageURL
}

func (a HueCycle) ActivityName() string {
	return "HueCycle"
}
