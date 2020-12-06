package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

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
	Session         *discordgo.Session
	Config          *Config
	Parser          *parsley.Parser
	PipelineManager *PipelineManager
	Storage         PersistenceBackend
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

	log.Debug().Msg("Creating persistence backend")
	var backend PersistenceBackend
	switch config.StorageType {
	case "consul":
		backend, err = NewConsulBackend(config)
		if err != nil {
			return nil, fmt.Errorf("error creating persistence backend: %w", err)
		}
	case "file":
		backend, err = NewFSBackend(config)
		if err != nil {
			return nil, fmt.Errorf("error creating persistence backend: %w", err)
		}
	default:
		return nil, fmt.Errorf("error creating persistence backend: %w", errors.New("unknown backend type"))
	}
	log.Debug().Msg("Persistence backend created")

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

	log.Debug().Msg("Creating pipeline manager")
	manager, err := NewPipelineManager(backend)
	if err != nil {
		return nil, fmt.Errorf("error creating pipeline manager: %w", err)
	}
	log.Debug().Msg("Pipeline manager created")

	log.Debug().Msg("Registering commands")
	parser.NewCommand("", "Magikify an image", _MagikCommand)
	parser.NewCommand("magik", "Magikify an image", _MagikCommand)
	parser.NewCommand("arcweld", "Arc-weld an image", _ArcweldCommand)
	parser.NewCommand("malt", "Malt an image", _MaltCommand)
	parser.NewCommand("help", "List available commands", _HelpCommand)
	parser.NewCommand("createpipeline", "Begin creation of a new command pipeline", _CreatePipelineCommand)
	parser.NewCommand("runpipeline", "Run a command pipeline", _RunPipelineCommand)
	parser.NewCommand("deletepipeline", "Delete a command pipeline", _DeletePipelineCommand)
	parser.NewCommand("savepipeline", "Save a pending pipeline", _SavePipelineCommand)
	parser.NewCommand("deepfry", "Deep-fry an image", _DeepfryCommand)
	log.Debug().Msg("Commands registered")

	Instance = &Borik{
		session,
		&config,
		parser,
		manager,
		backend,
	}
	log.Debug().Msg("Borik instance created")

	return Instance, nil
}

// TypingIndicator invokes a typing indicator in the channel of a message
func TypingIndicator(message *discordgo.MessageCreate) func() {
	stopTyping := Schedule(
		func() {
			log.Debug().Str("channel", message.ChannelID).Msg("Invoking typing indicator in channel")
			err := Instance.Session.ChannelTyping(message.ChannelID)
			if err != nil {
				log.Error().Err(err).Msg("Error while attempting invoke typing indicator in channel")
				return
			}
		},
		5*time.Second,
	)
	return func() {
		stopTyping <- true
	}
}

// PrepareAndInvokeOperation downloads the image pulled from the message, invokes the given operation with said image, and posts the image in the channel of the message that invoked it
func PrepareAndInvokeOperation(message *discordgo.MessageCreate, imageURL string, operation func([]byte, io.Writer) error) {
	srcBytes, err := DownloadImage(imageURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}
	destBuffer := new(bytes.Buffer)

	log.Debug().Msg("Beginning processing image")
	err = operation(srcBytes, destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process image")
		return
	}

	log.Debug().Msg("Image processed, uploading result")
	_, err = Instance.Session.ChannelFileSend(message.ChannelID, "test.jpeg", destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		_, err = Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
	}
}
