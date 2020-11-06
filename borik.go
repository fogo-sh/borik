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
	"github.com/saturn-sh/borik/bot"
)

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
	borik, err := bot.New()
	if err != nil {
		fmt.Printf("Error creating Borik instance: %s\n", err.Error())
		return
	}

	borik.Session.AddHandler(messageCreate)

	borik.Session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	err = borik.Session.Open()
	if err != nil {
		fmt.Printf("Error opening connection: %s\n", err.Error())
		return
	}

	fmt.Println("Borik is now running, press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	borik.Session.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	config := bot.Instance.Config
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
