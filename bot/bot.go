package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/shlex"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config represents the config that Borik will use to run
type Config struct {
	Prefix   string        `default:"borik!"`
	Token    string        `required:"true"`
	LogLevel zerolog.Level `default:"1" split_words:"true"`
}

// Command represents an individual command.
type Command struct {
	Name        string
	Description string
	Handler     func(event *discordgo.MessageCreate, args interface{})
}

// MarshalJSON marshals a command to json, omitting the handler function.
func (command Command) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		command.Name, command.Description,
	})
}

// Borik represents an individual instance of Borik
type Borik struct {
	Session  *discordgo.Session
	Config   *Config
	Commands []Command
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
	log.Debug().Msg("Discord session created")

	session.AddHandler(messageCreate)
	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	Instance = &Borik{
		session,
		&config,
		[]Command{
			{
				Name:        "magik",
				Description: "Magikify an image",
				Handler:     magikCommand,
			},
		},
	}

	log.Debug().Msg("Borik instance created")

	return Instance, nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	_ParseCommand(m)
}

func magikCommand(message *discordgo.MessageCreate, args interface{}) {
	argsSlice := args.([]string)
	imageURL := ""
	if len(argsSlice) <= 1 {
		var err error
		imageURL, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
		}
	} else {
		imageURL = argsSlice[1]
	}

	srcBytes, err := DownloadImage(imageURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}
	destBuffer := new(bytes.Buffer)

	log.Debug().Msg("Beginning processing image")
	err = Magik(srcBytes, destBuffer)
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

// _ParseCommand parses a message for commands, running the resulting command if found.
func _ParseCommand(message *discordgo.MessageCreate) error {
	log.Debug().Str("message", message.Content).Msg("Parsing command")
	args, err := shlex.Split(message.Content)
	if err != nil {
		log.Error().Err(err).Msg("Error processing message for commands")
		return fmt.Errorf("error processing message text: %w", err)
	}

	command := args[0]
	if !strings.HasPrefix(command, Instance.Config.Prefix) {
		return nil
	}
	command = strings.TrimPrefix(command, Instance.Config.Prefix)

	for _, commandObj := range Instance.Commands {
		if commandObj.Name == command {
			log.Debug().Interface("command", commandObj).Msg("Found command in message")
			commandObj.Handler(message, args)
		}
	}

	return nil
}
