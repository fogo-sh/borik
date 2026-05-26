package args

type PresidentsFrame struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a PresidentsFrame) GetImageURL() string {
	return a.ImageURL
}

func (a PresidentsFrame) ActivityName() string {
	return "PresidentsFrame"
}

type Heritage struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Heritage) GetImageURL() string {
	return a.ImageURL
}

func (a Heritage) ActivityName() string {
	return "Heritage"
}

type Shinji struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Shinji) GetImageURL() string {
	return a.ImageURL
}

func (a Shinji) ActivityName() string {
	return "Shinji"
}
