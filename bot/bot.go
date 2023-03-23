package bot

import (
	"context"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
	"github.com/nint8835/parsley"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc/credentials"
)

// PersistenceBackend represents a generic backend capable of persisting data.
type PersistenceBackend interface {
	Get(string, interface{}) error
	Put(string, interface{}) error
}

// Config represents the config that Borik will use to run
type Config struct {
	Prefix   string        `default:"borik!"`
	Token    string        `required:"true"`
	LogLevel zerolog.Level `default:"1" split_words:"true"`

	HoneycombToken   string `default:"" split_words:"true"`
	HoneycombDataset string `default:"" split_words:"true"`
}

// Borik represents an individual instance of Borik
type Borik struct {
	Session *discordgo.Session
	Config  *Config
	Parser  *parsley.Parser
	Trace   *sdktrace.TracerProvider
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
	_ = parser.NewCommand("", "Magikify an image.", MakeImageOpCommand(Magik, "magik"))
	_ = parser.NewCommand("magik", "Magikify an image.", MakeImageOpCommand(Magik, "magik"))
	_ = parser.NewCommand("lagik", "Lagikify an image.", MakeImageOpCommand(Lagik, "lagik"))
	_ = parser.NewCommand("gmagik", "Repeatedly magikify an image.", MakeImageOpCommand(Gmagik, "gmagik"))
	_ = parser.NewCommand("arcweld", "Arc-weld an image.", MakeImageOpCommand(Arcweld, "arcweld"))
	_ = parser.NewCommand("malt", "Malt an image.", MakeImageOpCommand(Malt, "malt"))
	_ = parser.NewCommand("help", "Get help for available commands.", HelpCommand)
	_ = parser.NewCommand("deepfry", "Deep-fry an image.", MakeImageOpCommand(Deepfry, "deepfry"))
	_ = parser.NewCommand("stevepoint", "Have Steve point at an image.", MakeImageOpCommand(StevePoint, "stevepoint"))
	_ = parser.NewCommand("mitchpoint", "Have Mitch point at an image.", MakeImageOpCommand(MitchPoint, "mitchpoint"))
	_ = parser.NewCommand("keenanthumb", "Have Keenan thumbs-up an image.", MakeImageOpCommand(KeenanThumb, "keenanthumb"))
	_ = parser.NewCommand("sidekeenan", "Have Keenan on the side of an image.", MakeImageOpCommand(SideKeenan, "sidekeenan"))
	_ = parser.NewCommand("divine", "Sever the divine light.", MakeImageOpCommand(Divine, "divine"))
	_ = parser.NewCommand("waaw", "Mirror the right half of an image.", MakeImageOpCommand(Waaw, "waaw"))
	_ = parser.NewCommand("haah", "Mirror the left half of an image.", MakeImageOpCommand(Haah, "haah"))
	_ = parser.NewCommand("woow", "Mirror the top half of an image.", MakeImageOpCommand(Woow, "woow"))
	_ = parser.NewCommand("hooh", "Mirror the bottom half of an image.", MakeImageOpCommand(Hooh, "hooh"))
	_ = parser.NewCommand("invert", "Invert the colours of an image.", MakeImageOpCommand(Invert, "invert"))
	_ = parser.NewCommand("otsu", "Apply a threshold to an image using Otsu's method.", MakeImageOpCommand(Otsu, "otsu"))
	_ = parser.NewCommand("rotate", "Rotate an image.", MakeImageOpCommand(Rotate, "rotate"))
	_ = parser.NewCommand("avatar", "Fetch the avatar for a user.", Avatar)
	_ = parser.NewCommand("sticker", "Fetch a sticker as an image.", Sticker)
	_ = parser.NewCommand("emoji", "Fetch an emoji as an image.", Emoji)
	_ = parser.NewCommand("resize", "Resize an image.", MakeImageOpCommand(Resize, "resize"))
	_ = parser.NewCommand("huecycle", "Create a GIF cycling the hue of an image.", MakeImageOpCommand(HueCycle, "huecycle"))
	_ = parser.NewCommand("modulate", "Modify the brightness, saturation, and hue of an image.", MakeImageOpCommand(Modulate, "modulate"))
	registerGraphicsFormatCommands(parser)

	log.Debug().Msg("Commands registered")

	traceResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("borik"),
	)

	var traceProvider *trace.TracerProvider

	if config.HoneycombToken != "" {
		log.Debug().Msg("Configuring Honeycomb exporter")
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint("api.honeycomb.io:443"),
			otlptracegrpc.WithHeaders(map[string]string{
				"x-honeycomb-team":    config.HoneycombToken,
				"x-honeycomb-dataset": config.HoneycombDataset,
			}),
			otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		}

		client := otlptracegrpc.NewClient(opts...)
		exporter, err := otlptrace.New(context.Background(), client)
		if err != nil {
			return nil, fmt.Errorf("error creating opentelemetry exporter: %w", err)
		}
		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(traceResource),
		)

		log.Debug().Msg("Honeycomb configured")
	} else {
		traceProvider = sdktrace.NewTracerProvider()
	}

	otel.SetTracerProvider(traceProvider)

	Instance = &Borik{
		session,
		&config,
		parser,
		traceProvider,
	}
	log.Debug().Msg("Borik instance created")

	return Instance, nil
}
