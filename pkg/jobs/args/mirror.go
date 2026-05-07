package args

type Waaw struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Waaw) GetImageURL() string {
	return a.ImageURL
}

func (a Waaw) ActivityName() string {
	return "Waaw"
}

type Haah struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Haah) GetImageURL() string {
	return a.ImageURL
}

func (a Haah) ActivityName() string {
	return "Haah"
}

type Woow struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Woow) GetImageURL() string {
	return a.ImageURL
}

func (a Woow) ActivityName() string {
	return "Woow"
}

type Hooh struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Hooh) GetImageURL() string {
	return a.ImageURL
}

func (a Hooh) ActivityName() string {
	return "Hooh"
}
