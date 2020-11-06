package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the config that Borik will use to run
type Config struct {
	Prefix string `default:"borik!"`
	Token  string `required:"true"`
}

// Borik represents an individual instance of Borik
type Borik struct {
	Session *discordgo.Session
	Config  *Config
}

// Instance is the current instance of Borik
var Instance *Borik

// New constructs a new instance of Borik.
func New() (*Borik, error) {
	var config Config
	err := envconfig.Process("borik", &config)
	if err != nil {
		return nil, fmt.Errorf("error loading Borik config: %w", err)
	}

	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating new Discord session: %w", err)
	}

	Instance = &Borik{session, &config}

	return Instance, nil
}
