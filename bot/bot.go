package bot

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
	"github.com/nint8835/parsley"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// PersistenceBackend represents a generic backend capable of persisting data.
type PersistenceBackend interface {
	Get(string, interface{}) error
	Put(string, interface{}) error
}

// Config represents the config that Borik will use to run
type Config struct {
	Prefix        string        `default:"borik!"`
	Token         string        `required:"true"`
	LogLevel      zerolog.Level `default:"1" split_words:"true"`
	StorageType   string        `default:"file" split_words:"true"`
	FilePath      string        `default:"backend" split_words:"true"`
	ConsulAddress string        `default:"" split_words:"true"`
}

// Borik represents an individual instance of Borik
type Borik struct {
	Session *discordgo.Session
	Config  *Config
	Parser  *parsley.Parser
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

	zerolog.SetGlobalLevel(config.LogLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Debug().Msg("Creating Discord session")
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating new Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	log.Debug().Msg("Discord session created")

	log.Debug().Msg("Creating command parser")
	parser := parsley.New(config.Prefix)
	parser.RegisterHandler(session)
	log.Debug().Msg("Parser created")

	log.Debug().Msg("Registering commands")
	_ = parser.NewCommand("", "Magikify an image.", MakeImageOpCommandV7(Magik))
	_ = parser.NewCommand("magik", "Magikify an image.", MakeImageOpCommandV7(Magik))
	_ = parser.NewCommand("lagik", "Lagikify an image.", MakeImageOpCommandV7(Lagik))
	_ = parser.NewCommand("gmagik", "Repeatedly magikify an image.", MakeImageOpCommandV7(Gmagik))
	_ = parser.NewCommand("arcweld", "Arc-weld an image.", MakeImageOpCommandV6(Arcweld))
	_ = parser.NewCommand("malt", "Malt an image.", MakeImageOpCommandV6(Malt))
	_ = parser.NewCommand("help", "Get help for available commands.", HelpCommand)
	_ = parser.NewCommand("deepfry", "Deep-fry an image.", MakeImageOpCommandV7(Deepfry))
	_ = parser.NewCommand("stevepoint", "Have Steve point at an image.", MakeImageOpCommandV7(StevePoint))
	_ = parser.NewCommand("mitchpoint", "Have Mitch point at an image.", MakeImageOpCommandV7(MitchPoint))
	_ = parser.NewCommand("divine", "Sever the divine light.", MakeImageOpCommandV6(Divine))
	_ = parser.NewCommand("waaw", "Mirror the right half of an image.", MakeImageOpCommandV7(Waaw))
	_ = parser.NewCommand("haah", "Mirror the left half of an image.", MakeImageOpCommandV7(Haah))
	_ = parser.NewCommand("woow", "Mirror the top half of an image.", MakeImageOpCommandV7(Woow))
	_ = parser.NewCommand("hooh", "Mirror the bottom half of an image.", MakeImageOpCommandV7(Hooh))
	_ = parser.NewCommand("transform", "Apply transformations to an image.", MakeImageOpCommandV6(Transform))
	_ = parser.NewCommand("invert", "Invert the colours of an image.", MakeImageOpCommandV6(Invert))
	_ = parser.NewCommand("otsu", "Apply a threshold to an image using Otsu's method.", MakeImageOpCommandV6(Otsu))
	_ = parser.NewCommand("rotate", "Rotate an image.", MakeImageOpCommandV6(Rotate))
	_ = parser.NewCommand("avatar", "Fetch the avatar for a user.", Avatar)
	_ = parser.NewCommand("sticker", "Fetch a sticker as an image.", Sticker)
	_ = parser.NewCommand("emoji", "Fetch an emoji as an image.", Emoji)
	registerGraphicsFormatCommands(parser)

	log.Debug().Msg("Commands registered")

	Instance = &Borik{
		session,
		&config,
		parser,
	}
	log.Debug().Msg("Borik instance created")

	return Instance, nil
}
