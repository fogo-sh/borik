package args

type JackPog struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a JackPog) GetImageURL() string  { return a.ImageURL }
func (a JackPog) ActivityName() string { return "JackPog" }

type SideKeenan struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a SideKeenan) GetImageURL() string  { return a.ImageURL }
func (a SideKeenan) ActivityName() string { return "SideKeenan" }

type KeenanThumb struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a KeenanThumb) GetImageURL() string  { return a.ImageURL }
func (a KeenanThumb) ActivityName() string { return "KeenanThumb" }

type MitchPoint struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a MitchPoint) GetImageURL() string  { return a.ImageURL }
func (a MitchPoint) ActivityName() string { return "MitchPoint" }

type StevePoint struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a StevePoint) GetImageURL() string  { return a.ImageURL }
func (a StevePoint) ActivityName() string { return "StevePoint" }

type AndrewPog struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a AndrewPog) GetImageURL() string  { return a.ImageURL }
func (a AndrewPog) ActivityName() string { return "AndrewPog" }

type MatlabKid struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a MatlabKid) GetImageURL() string  { return a.ImageURL }
func (a MatlabKid) ActivityName() string { return "MatlabKid" }

type NatalieClimb struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a NatalieClimb) GetImageURL() string  { return a.ImageURL }
func (a NatalieClimb) ActivityName() string { return "NatalieClimb" }

type DennyStanding struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip the overlay horizontally."`
	VFlip    bool   `default:"false" description:"Flip the overlay vertically."`
}

func (a DennyStanding) GetImageURL() string  { return a.ImageURL }
func (a DennyStanding) ActivityName() string { return "DennyStanding" }
