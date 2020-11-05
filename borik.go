package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/esimov/caire"
	"github.com/kelseyhightower/envconfig"
	"github.com/saturn-sh/borik/bot"
)

// Config represents the config that borik will use to run
type Config struct {
	Prefix string `default:"borik!"`
	Token  string `required:"true"`
}

var config Config

func gik(in io.Reader, out io.Writer) {
	p := &caire.Processor{
		// TODO calculate width / height
		// NewWidth: 512,
		// NewHeight: 512,
	}

	if err := p.Process(in, out); err != nil {
		fmt.Printf("Error rescaling image: %s", err.Error())
	}
}

func main() {
	err := envconfig.Process("borik", &config)
	if err != nil {
		fmt.Printf("error loading config: %s\n", err)
		return
	}

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("borik is now running, press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, config.Prefix) {
		return
	}

	imageURI, err := bot.ImageURIFromCommand(s, m, config.Prefix)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	fmt.Println("found imageURI to borik: ", imageURI)

	// s.ChannelMessageSend(m.ChannelID, imageURI)
}
