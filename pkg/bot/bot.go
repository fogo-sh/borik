package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/rs/zerolog/log"

	configPkg "github.com/fogo-sh/borik/pkg/config"
)

// Bot represents an individual instance of Borik
type Bot struct {
	session      *discordgo.Session
	openAiClient openai.Client
	config       *configPkg.Config
	parser       *parsley.Parser
	quitChan     chan struct{}
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening discord session: %w", err)
	}

	<-b.quitChan

	err = b.session.Close()
	if err != nil {
		return fmt.Errorf("error closing discord session: %w", err)
	}

	return nil
}

func (b *Bot) Stop() {
	b.quitChan <- struct{}{}
}

// Instance is the current instance of Borik
var Instance *Bot

// New constructs a new instance of Borik.
func New() (*Bot, error) {
	config := configPkg.Instance

	openAiClient := openai.NewClient(
		option.WithBaseURL(config.OpenaiBaseUrl),
		option.WithAPIKey(config.OpenaiApiKey),
	)

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
	_ = parser.NewCommand("lagik", "Lagikify an image.", MakeImageOpCommand(Lagik))
	_ = parser.NewCommand("gmagik", "Repeatedly magikify an image.", MakeImageOpCommand(Gmagik))
	_ = parser.NewCommand("arcweld", "Arc-weld an image.", MakeImageOpCommand(Arcweld))
	_ = parser.NewCommand("malt", "Malt an image.", MakeImageOpCommand(Malt))
	_ = parser.NewCommand("help", "Get help for available commands.", HelpCommand)
	_ = parser.NewCommand("deepfry", "Deep-fry an image.", MakeImageOpCommand(Deepfry))
	_ = parser.NewCommand("divine", "Sever the divine light.", MakeImageOpCommand(Divine))
	_ = parser.NewCommand("waaw", "Mirror the right half of an image.", MakeImageOpCommand(Waaw))
	_ = parser.NewCommand("haah", "Mirror the left half of an image.", MakeImageOpCommand(Haah))
	_ = parser.NewCommand("woow", "Mirror the top half of an image.", MakeImageOpCommand(Woow))
	_ = parser.NewCommand("hooh", "Mirror the bottom half of an image.", MakeImageOpCommand(Hooh))
	_ = parser.NewCommand("invert", "Invert the colours of an image.", MakeImageOpCommand(Invert))
	_ = parser.NewCommand("otsu", "Apply a threshold to an image using Otsu's method.", MakeImageOpCommand(Otsu))
	_ = parser.NewCommand("rotate", "Rotate an image.", MakeImageOpCommand(Rotate))
	_ = parser.NewCommand("avatar", "Fetch the avatar for a user.", Avatar)
	_ = parser.NewCommand("sticker", "Fetch a sticker as an image.", Sticker)
	_ = parser.NewCommand("emoji", "Fetch an emoji as an image.", Emoji)
	_ = parser.NewCommand("resize", "Resize an image.", MakeImageOpCommand(Resize))
	_ = parser.NewCommand("huecycle", "Create a GIF cycling the hue of an image.", MakeImageOpCommand(HueCycle))
	_ = parser.NewCommand("modulate", "Modify the brightness, saturation, and hue of an image.", MakeImageOpCommand(Modulate))
	_ = parser.NewCommand("presidentsframe", "Apply the President's Frame to an image", MakeImageOpCommand(PresidentsFrame))
	_ = parser.NewCommand("meme", "Add meme text to an image.", MakeImageOpCommand(Meme))
	registerGraphicsFormatCommands(parser)
	registerOverlayCommands(parser)

	if config.OpenaiApiKey != "" {
		log.Debug().Msg("Registering OpenAI commands")
		_ = parser.NewCommand("imagegen", "Generate an image from a prompt.", ImageGen)
		_ = parser.NewCommand("imageedit", "Edit an image based on a prompt.", ImageEditCommand)
		_ = parser.NewCommand("loopedit", "Repeatedly edit an image based on a prompt.", LoopEditCommand)
	} else {
		log.Warn().Msg("OpenAI API key not set; skipping registration of OpenAI commands")
	}

	log.Debug().Msg("Commands registered")

	Instance = &Bot{
		session,
		openAiClient,
		config,
		parser,
		make(chan struct{}),
	}

	return Instance, nil
}
