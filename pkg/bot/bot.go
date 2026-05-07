package bot

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/client"
	"pkg.nit.so/switchboard"

	configPkg "github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/logging"
)

// Bot represents an individual instance of Borik
type Bot struct {
	session        *discordgo.Session
	openAiClient   openai.Client
	config         *configPkg.Config
	textParser     *parsley.Parser
	slashParser    *switchboard.Switchboard
	temporalClient client.Client
	quitChan       chan struct{}
}

func (b *Bot) Start() error {
	defer b.temporalClient.Close()

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
	textHandler  any
	slashHandler any
	enabled      func(*configPkg.Config) bool
}

var commands = []Command{
	{
		name:         "magik",
		slashAliases: []string{"borik"},
		description:  "Magikify an image.",
		textHandler:  MakeWorkflowTextCommand[args.Magik](),
		slashHandler: MakeWorkflowSlashCommand[args.Magik](),
	},
	{
		name:         "lagik",
		description:  "Lagikify an image.",
		textHandler:  MakeWorkflowTextCommand[args.Lagik](),
		slashHandler: MakeWorkflowSlashCommand[args.Lagik](),
	},
	{
		name:         "gmagik",
		description:  "Repeatedly magikify an image.",
		textHandler:  MakeWorkflowTextCommand[args.Gmagik](),
		slashHandler: MakeWorkflowSlashCommand[args.Gmagik](),
	},
	{
		name:         "arcweld",
		description:  "Arc-weld an image.",
		textHandler:  MakeWorkflowTextCommand[args.Arcweld](),
		slashHandler: MakeWorkflowSlashCommand[args.Arcweld](),
	},
	{
		name:         "malt",
		description:  "Malt an image.",
		textHandler:  MakeWorkflowTextCommand[args.Malt](),
		slashHandler: MakeWorkflowSlashCommand[args.Malt](),
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
		textHandler:  MakeWorkflowTextCommand[args.Deepfry](),
		slashHandler: MakeWorkflowSlashCommand[args.Deepfry](),
	},
	{
		name:         "divine",
		description:  "Sever the divine light.",
		textHandler:  MakeWorkflowTextCommand[args.Divine](),
		slashHandler: MakeWorkflowSlashCommand[args.Divine](),
	},
	// {
	// 	name:         "waaw",
	// 	description:  "Mirror the right half of an image.",
	// 	textHandler:  MakeImageOpTextCommand(Waaw),
	// 	slashHandler: MakeImageOpSlashCommand(Waaw),
	// },
	// {
	// 	name:         "haah",
	// 	description:  "Mirror the left half of an image.",
	// 	textHandler:  MakeImageOpTextCommand(Haah),
	// 	slashHandler: MakeImageOpSlashCommand(Haah),
	// },
	// {
	// 	name:         "woow",
	// 	description:  "Mirror the top half of an image.",
	// 	textHandler:  MakeImageOpTextCommand(Woow),
	// 	slashHandler: MakeImageOpSlashCommand(Woow),
	// },
	// {
	// 	name:         "hooh",
	// 	description:  "Mirror the bottom half of an image.",
	// 	textHandler:  MakeImageOpTextCommand(Hooh),
	// 	slashHandler: MakeImageOpSlashCommand(Hooh),
	// },
	// {
	// 	name:         "invert",
	// 	description:  "Invert the colours of an image.",
	// 	textHandler:  MakeImageOpTextCommand(Invert),
	// 	slashHandler: MakeImageOpSlashCommand(Invert),
	// },
	// {
	// 	name:         "otsu",
	// 	description:  "Apply a threshold to an image using Otsu's method.",
	// 	textHandler:  MakeImageOpTextCommand(Otsu),
	// 	slashHandler: MakeImageOpSlashCommand(Otsu),
	// },
	// {
	// 	name:         "rotate",
	// 	description:  "Rotate an image.",
	// 	textHandler:  MakeImageOpTextCommand(Rotate),
	// 	slashHandler: MakeImageOpSlashCommand(Rotate),
	// },
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
	// {
	// 	name:         "resize",
	// 	description:  "Resize an image.",
	// 	textHandler:  MakeImageOpTextCommand(Resize),
	// 	slashHandler: MakeImageOpSlashCommand(Resize),
	// },
	// {
	// 	name:         "huecycle",
	// 	description:  "Create a GIF cycling the hue of an image.",
	// 	textHandler:  MakeImageOpTextCommand(HueCycle),
	// 	slashHandler: MakeImageOpSlashCommand(HueCycle),
	// },
	{
		name:         "gif",
		description:  "Convert a video to a GIF.",
		textHandler:  GifTextCommand,
		slashHandler: GifSlashCommand,
	},
	// {
	// 	name:         "modulate",
	// 	description:  "Modify the brightness, saturation, and hue of an image.",
	// 	textHandler:  MakeImageOpTextCommand(Modulate),
	// 	slashHandler: MakeImageOpSlashCommand(Modulate),
	// },
	// {
	// 	name:         "meme",
	// 	description:  "Add meme text to an image.",
	// 	textHandler:  MakeImageOpTextCommand(Meme),
	// 	slashHandler: MakeImageOpSlashCommand(Meme),
	// },
	// {
	// 	name:         "hdr",
	// 	description:  "Apply aggressive HDR color boosting to an image.",
	// 	textHandler:  MakeImageOpTextCommand(Hdr),
	// 	slashHandler: MakeImageOpSlashCommand(Hdr),
	// },
	{
		name:         "aigen",
		description:  "Generate an image from a prompt.",
		textHandler:  ImageGenTextCommand,
		slashHandler: ImageGenSlashCommand,
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "aiedit",
		description:  "Edit an image based on a prompt.",
		textHandler:  MakeAIImageOpTextCommand(ImageEdit),
		slashHandler: MakeAIImageOpSlashCommand(ImageEdit),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "ailoopedit",
		description:  "Repeatedly edit an image based on a prompt.",
		textHandler:  MakeAIImageOpTextCommand(LoopEdit),
		slashHandler: MakeAIImageOpSlashCommand(LoopEdit),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "aiflipflop",
		description:  "Flip-flop between two images, editing each based on a prompt.",
		textHandler:  MakeAIImageOpTextCommand(FlipFlop),
		slashHandler: MakeAIImageOpSlashCommand(FlipFlop),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "aizoom",
		description:  "Zoom out from an image.",
		textHandler:  MakeAIImageOpTextCommand(AiZoom),
		slashHandler: MakeAIImageOpSlashCommand(AiZoom),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
	{
		name:         "ailoopzoom",
		description:  "Repeatedly zoom out from an image.",
		textHandler:  MakeAIImageOpTextCommand(AiLoopZoom),
		slashHandler: MakeAIImageOpSlashCommand(AiLoopZoom),
		enabled:      func(c *configPkg.Config) bool { return c.OpenaiApiKey != "" },
	},
}

// New constructs a new instance of Borik.
func New() (*Bot, error) {
	config := configPkg.Instance
	if strings.TrimSpace(config.Token) == "" {
		return nil, fmt.Errorf("Discord bot token must be set")
	}

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
	textParser := parsley.New(config.Prefixes...)
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

	c, err := client.Dial(client.Options{
		Namespace: config.TemporalNamespace,
		Logger:    logging.NewTemporalLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating temporal client: %w", err)
	}

	Instance = &Bot{
		session:        session,
		openAiClient:   openAiClient,
		config:         config,
		textParser:     textParser,
		slashParser:    slashParser,
		temporalClient: c,
		quitChan:       make(chan struct{}),
	}

	slashGuildId := config.GuildId
	if config.RegisterSlashCommandsGlobally {
		slashGuildId = ""
	}

	log.Debug().Msg("Registering commands")

	if config.OpenaiApiKey == "" {
		log.Warn().Msg("OpenAI API key not set; skipping registration of OpenAI commands")
	}

	_ = textParser.NewCommand("", "Magikify an image.", MakeWorkflowTextCommand[args.Magik]())

	allCommands := slices.Concat(
		commands,
		generateGraphicsFormatCommands(),
		generateFrameCommands(),
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

	return Instance, nil
}
