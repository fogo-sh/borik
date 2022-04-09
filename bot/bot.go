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
	_ = parser.NewCommand("", "Magikify an image.", MakeImageOpCommand(Magik))
	_ = parser.NewCommand("magik", "Magikify an image.", MakeImageOpCommand(Magik))
	_ = parser.NewCommand("gmagik", "Repeatedly magikify an image.", MakeImageOpCommand(Gmagik))
	_ = parser.NewCommand("arcweld", "Arc-weld an image.", MakeImageOpCommand(Arcweld))
	_ = parser.NewCommand("malt", "Malt an image.", MakeImageOpCommand(Malt))
	_ = parser.NewCommand("help", "Get help for available commands.", HelpCommand)
	_ = parser.NewCommand("deepfry", "Deep-fry an image.", MakeImageOpCommand(Deepfry))
	_ = parser.NewCommand("stevepoint", "Have Steve point at an image.", MakeImageOpCommand(StevePoint))
	_ = parser.NewCommand("divine", "Sever the divine light.", MakeImageOpCommand(Divine))
	_ = parser.NewCommand("waaw", "Mirror the right half of an image.", MakeImageOpCommand(Waaw))
	_ = parser.NewCommand("haah", "Mirror the left half of an image.", MakeImageOpCommand(Haah))
	_ = parser.NewCommand("woow", "Mirror the top half of an image.", MakeImageOpCommand(Woow))
	_ = parser.NewCommand("hooh", "Mirror the bottom half of an image.", MakeImageOpCommand(Hooh))
	_ = parser.NewCommand("transform", "Apply transformations to an image.", MakeImageOpCommand(Transform))
	log.Debug().Msg("Commands registered")

	Instance = &Borik{
		session,
		&config,
		parser,
	}
	log.Debug().Msg("Borik instance created")

	return Instance, nil
}
