package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/esimov/caire"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the config that borik will use to run
type Config struct {
	Prefix string `default:"borik!"`
	Token  string `required:"true"`
}

var config Config

func downloadFile(filepath string, url string) (err error) {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

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

func imageURIFromMessage(m *discordgo.Message) *string {
	if len(m.Embeds) == 1 {
		embed := m.Embeds[0]

		if embed.Type == "Image" {
			return &embed.URL
		}
	}

	if len(m.Attachments) == 1 {
		attachment := m.Attachments[0]
		return &attachment.URL
	}

	return nil
}

func imageURIFromCommand(s *discordgo.Session, m *discordgo.MessageCreate) (*string, error) {
	argument := strings.TrimSpace(strings.TrimPrefix(m.Content, config.Prefix))

	if argument != "" {
		return &argument, nil
	}

	imageURI := imageURIFromMessage(m.Message)
	if imageURI != nil {
		return imageURI, nil
	}

	messages, err := s.ChannelMessages(m.ChannelID, 20, m.ID, "", "")
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		imageURI := imageURIFromMessage(message)
		if imageURI != nil {
			return imageURI, nil
		}
	}

	return nil, errors.New("no image found from message")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, config.Prefix) {
		return
	}

	imageURI, err := imageURIFromCommand(s, m)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	fmt.Println("found imageURI to borik: ", imageURI)

	// s.ChannelMessageSend(m.ChannelID, imageURI)
}
