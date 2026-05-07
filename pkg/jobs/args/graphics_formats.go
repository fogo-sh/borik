package args

type EGA struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Dither   bool   `default:"false" description:"Whether the final image should be dithered."`
}

func (a EGA) GetImageURL() string {
	return a.ImageURL
}

func (a EGA) ActivityName() string {
	return "EGA"
}

type TempleOS struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Dither   bool   `default:"false" description:"Whether the final image should be dithered."`
}

func (a TempleOS) GetImageURL() string {
	return a.ImageURL
}

func (a TempleOS) ActivityName() string {
	return "TempleOS"
}

type CGA struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Dither   bool   `default:"false" description:"Whether the final image should be dithered."`
}

func (a CGA) GetImageURL() string {
	return a.ImageURL
}

func (a CGA) ActivityName() string {
	return "CGA"
}
