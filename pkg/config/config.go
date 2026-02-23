package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Prefixes is a custom type for command prefixes, split on "|" to allow
// commas and other special characters as prefix values.
// Example: BORIK_PREFIXES="borik!|,"
type Prefixes []string

func (p *Prefixes) Decode(value string) error {
	*p = strings.Split(value, "|")
	return nil
}

// Config represents the config that Borik will use to run
type Config struct {
	Prefixes Prefixes `default:"borik!"`
	Token    string   `required:"true"`
	LogLevel string   `default:"info" split_words:"true"`

	OpenaiBaseUrl        string `default:"https://llm.ops.bootleg.technology/v1" split_words:"true"`
	OpenaiApiKey         string `default:"" split_words:"true"`
	OpenaiImageGenModel  string `default:"flux-2-klein-4b" split_words:"true"`
	OpenaiImageEditModel string `default:"flux-2-klein-4b" split_words:"true"`
}

var Instance *Config

func Load() error {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Msg("Error loading .env file")
	}

	var newConfig Config
	err = envconfig.Process("borik", &newConfig)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(newConfig.LogLevel)
	if err != nil {
		return fmt.Errorf("error parsing log level: %w", err)
	}
	zerolog.SetGlobalLevel(logLevel)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	Instance = &newConfig

	return nil
}
