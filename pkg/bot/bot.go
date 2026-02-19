package bot

import (
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/rs/zerolog/log"
	"pkg.nit.so/switchboard"

	configPkg "github.com/fogo-sh/borik/pkg/config"
)

// Bot represents an individual instance of Borik
type Bot struct {
	session      *discordgo.Session
	openAiClient openai.Client
	config       *configPkg.Config
	textParser   *parsley.Parser
	slashParser  *switchboard.Switchboard
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

type Command struct {
	name         string
	aliases      []string
	slashAliases []string
	description  string
	textHandler  interface{}
	slashHandler interface{}
	enabled      func(*configPkg.Config) bool
}

var commands = []Command{
	{
		name:         "magik",
		slashAliases: []string{"borik"},
		description:  "Magikify an image.",
		textHandler:  MakeImageOpTextCommand(Magik),
		slashHandler: MakeImageOpSlashCommand(Magik),
	},
	{
		name:         "lagik",
		description:  "Lagikify an image.",
		textHandler:  MakeImageOpTextCommand(Lagik),
		slashHandler: MakeImageOpSlashCommand(Lagik),
	},
	{
		name:         "gmagik",
		description:  "Repeatedly magikify an image.",
		textHandler:  MakeImageOpTextCommand(Gmagik),
		slashHandler: MakeImageOpSlashCommand(Gmagik),
	},
	{
		name:         "arcweld",
		description:  "Arc-weld an image.",
		textHandler:  MakeImageOpTextCommand(Arcweld),
		slashHandler: MakeImageOpSlashCommand(Arcweld),
	},
	{
		name:         "malt",
		description:  "Malt an image.",
		textHandler:  MakeImageOpTextCommand(Malt),
		slashHandler: MakeImageOpSlashCommand(Malt),
	},
	{
		name:         "help",
		description:  "Get help for available commands.",
		textHandler:  HelpCommand,
		slashHandler: HelpSlashCommand,
	},
	{
		name:         "deepfry",
		description:  "Deep-fry an image.",
		textHandler:  MakeImageOpTextCommand(Deepfry),
		slashHandler: MakeImageOpSlashCommand(Deepfry),
	},
	{
		name:         "divine",
		description:  "Sever the divine light.",
		textHandler:  MakeImageOpTextCommand(Divine),
		slashHandler: MakeImageOpSlashCommand(Divine),
	},
	{
		name:         "waaw",
		description:  "Mirror the right half of an image.",
		textHandler:  MakeImageOpTextCommand(Waaw),
		slashHandler: MakeImageOpSlashCommand(Waaw),
	},
	{
		name:         "haah",
		description:  "Mirror the left half of an image.",
		textHandler:  MakeImageOpTextCommand(Haah),
		slashHandler: MakeImageOpSlashCommand(Haah),
	},
	{
		name:         "woow",
		description:  "Mirror the top half of an image.",
		textHandler:  MakeImageOpTextCommand(Woow),
		slashHandler: MakeImageOpSlashCommand(Woow),
	},
	{
		name:         "hooh",
		description:  "Mirror the bottom half of an image.",
		textHandler:  MakeImageOpTextCommand(Hooh),
		slashHandler: MakeImageOpSlashCommand(Hooh),
	},
	{
		name:         "invert",
		description:  "Invert the colours of an image.",
		textHandler:  MakeImageOpTextCommand(Invert),
		slashHandler: MakeImageOpSlashCommand(Invert),
	},
	{
		name:         "otsu",
		description:  "Apply a threshold to an image using Otsu's method.",
		textHandler:  MakeImageOpTextCommand(Otsu),
		slashHandler: MakeImageOpSlashCommand(Otsu),
	},
	{
		name:         "rotate",
		description:  "Rotate an image.",
		textHandler:  MakeImageOpTextCommand(Rotate),
		slashHandler: MakeImageOpSlashCommand(Rotate),
	},
	{
		name:         "avatar",
		description:  "Fetch the avatar for a user.",
		textHandler:  Avatar,
		slashHandler: AvatarSlashCommand,
	},
	{
		name:         "sticker",
		description:  "Fetch a sticker as an image.",
		textHandler:  Sticker,
		slashHandler: nil,
	},
	{
		name:         "emoji",
		description:  "Fetch an emoji as an image.",
		textHandler:  Emoji,
		slashHandler: nil,
	},
	{
		name:         "resize",
		description:  "Resize an image.",
		textHandler:  MakeImageOpTextCommand(Resize),
		slashHandler: MakeImageOpSlashCommand(Resize),
	},
	{
		name:         "huecycle",
		description:  "Create a GIF cycling the hue of an image.",
		textHandler:  MakeImageOpTextCommand(HueCycle),
		slashHandler: MakeImageOpSlashCommand(HueCycle),
	},
	{
		name:         "modulate",
		description:  "Modify the brightness, saturation, and hue of an image.",
		textHandler:  MakeImageOpTextCommand(Modulate),
		slashHandler: MakeImageOpSlashCommand(Modulate),
	},
	{
		name:         "presidentsframe",
		description:  "Apply the President's Frame to an image",
		textHandler:  MakeImageOpTextCommand(PresidentsFrame),
		slashHandler: MakeImageOpSlashCommand(PresidentsFrame),
	},
	{
		name:         "meme",
		description:  "Add meme text to an image.",
		textHandler:  MakeImageOpTextCommand(Meme),
		slashHandler: MakeImageOpSlashCommand(Meme),
	},
	{
		name:         "imagegen",
		description:  "Generate an image from a prompt.",
		textHandler:  ImageGenTextCommand,
		slashHandler: ImageGenSlashCommand,
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "imageedit",
		description:  "Edit an image based on a prompt.",
		textHandler:  MakeImageOpTextCommand(ImageEdit),
		slashHandler: MakeImageOpSlashCommand(ImageEdit),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "loopedit",
		description:  "Repeatedly edit an image based on a prompt.",
		textHandler:  MakeImageOpTextCommand(LoopEdit),
		slashHandler: MakeImageOpSlashCommand(LoopEdit),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
}

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

	log.Debug().Msg("Creating text command parser")
	textParser := parsley.New(config.Prefix)
	textParser.RegisterHandler(session)
	log.Debug().Msg("Text command parser created")

	slashEnabled := config.GuildId != "" || config.RegisterSlashCommandsGlobally
	var slashParser *switchboard.Switchboard
	if slashEnabled {
		if config.AppId == "" {
			return nil, fmt.Errorf("app ID must be set when slash commands are enabled")
		}

		log.Debug().Msg("Creating slash command parser")
		slashParser = &switchboard.Switchboard{}
		session.AddHandler(slashParser.HandleInteractionCreate)

		if config.RegisterSlashCommandsGlobally {
			log.Info().Msg("Slash commands will be registered globally")
		} else {
			log.Info().Str("guild_id", config.GuildId).Msg("Slash commands will be registered for guild")
		}

		log.Debug().Msg("Slash command parser created")
	} else {
		log.Warn().Msg("Guild ID not set and global registration disabled; skipping registration of slash commands")
	}

	slashGuildId := config.GuildId
	if config.RegisterSlashCommandsGlobally {
		slashGuildId = ""
	}

	log.Debug().Msg("Registering commands")

	if config.OpenaiApiKey == "" {
		log.Warn().Msg("OpenAI API key not set; skipping registration of OpenAI commands")
	}

	_ = textParser.NewCommand("", "Magikify an image.", MakeImageOpTextCommand(Magik))

	allCommands := slices.Concat(
		commands,
		generateGraphicsFormatCommands(),
		generateOverlayCommands(),
	)

	for _, command := range allCommands {
		if command.enabled != nil && !command.enabled(config) {
			log.Debug().Str("command", command.name).Msg("Skipping disabled command")
			continue
		}

		_ = textParser.NewCommand(
			command.name,
			command.description,
			command.textHandler,
		)

		if slashParser != nil && command.slashHandler != nil {
			_ = slashParser.AddCommand(&switchboard.Command{
				Name:        command.name,
				Description: command.description,
				Handler:     command.slashHandler,
				GuildID:     slashGuildId,
			})
		}

		for _, alias := range command.aliases {
			_ = textParser.NewCommand(
				alias,
				command.description,
				command.textHandler,
			)

			if slashParser != nil && command.slashHandler != nil {
				_ = slashParser.AddCommand(&switchboard.Command{
					Name:        alias,
					Description: command.description,
					Handler:     command.slashHandler,
					GuildID:     slashGuildId,
				})
			}
		}

		if slashParser != nil && command.slashHandler != nil {
			for _, alias := range command.slashAliases {
				_ = slashParser.AddCommand(&switchboard.Command{
					Name:        alias,
					Description: command.description,
					Handler:     command.slashHandler,
					GuildID:     slashGuildId,
				})
			}
		}
	}

	if slashParser != nil {
		err = slashParser.SyncCommands(session, config.AppId)
		if err != nil {
			return nil, fmt.Errorf("error syncing commands: %w", err)
		}
	}

	log.Debug().Msg("Commands registered")

	Instance = &Bot{
		session,
		openAiClient,
		config,
		textParser,
		slashParser,
		make(chan struct{}),
	}

	return Instance, nil
}
