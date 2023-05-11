package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"

	"github.com/fogo-sh/borik/pkg/config"
)

type Bot struct {
	session  *discordgo.Session
	parser   *parsley.Parser
	quitChan chan struct{}
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

func New() (*Bot, error) {
	bot := &Bot{
		quitChan: make(chan struct{}),
	}

	parser := parsley.New(config.Instance.Prefix)
	bot.parser = parser

	session, err := discordgo.New("Bot " + config.Instance.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating new Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuildMessages
	bot.session = session

	_ = parser.NewCommand("help", "Get help for available commands.", bot.helpCommand)

	parser.RegisterHandler(session)

	return bot, nil
}
