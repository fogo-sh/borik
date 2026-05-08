package bot

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type OperationContext struct {
	Session     *discordgo.Session
	Message     *discordgo.MessageCreate
	Interaction *discordgo.InteractionCreate
	deferred    bool
}

func NewOperationContextFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) *OperationContext {
	return &OperationContext{
		Session: session,
		Message: message,
	}
}

func NewOperationContextFromInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) *OperationContext {
	return &OperationContext{
		Session:     session,
		Interaction: interaction,
	}
}

func (ctx *OperationContext) GetSourceID() string {
	if ctx.Message != nil {
		return ctx.Message.ID
	} else if ctx.Interaction != nil {
		return ctx.Interaction.ID
	}
	return ""
}

func (ctx *OperationContext) GetUserID() string {
	if ctx.Message != nil {
		return ctx.Message.Author.ID
	} else if ctx.Interaction != nil {
		if ctx.Interaction.Member != nil {
			return ctx.Interaction.Member.User.ID
		}
		if ctx.Interaction.User != nil {
			return ctx.Interaction.User.ID
		}
	}
	return ""
}

// DeferResponse defers the interaction response for long-running operations.
// This is a no-op for message-based commands (typing indicator handles that case).
func (ctx *OperationContext) DeferResponse() error {
	if ctx.Interaction == nil {
		return nil
	}
	err := ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to send deferred interaction response: %w", err)
	}
	ctx.deferred = true
	return nil
}

// SendText sends a plain text message.
func (ctx *OperationContext) SendText(content string) error {
	if ctx.Message != nil {
		_, err := ctx.Session.ChannelMessageSendReply(ctx.Message.ChannelID, content, ctx.Message.Reference())
		return err
	}
	if ctx.deferred {
		_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// SendEmbed sends an embed message.
func (ctx *OperationContext) SendEmbed(embed *discordgo.MessageEmbed) error {
	if ctx.Message != nil {
		_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, embed)
		return err
	}
	if ctx.deferred {
		_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		return err
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// SendFiles sends one or more file attachments.
func (ctx *OperationContext) SendFiles(files []*discordgo.File) error {
	if ctx.Message != nil {
		_, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, &discordgo.MessageSend{
			Reference: ctx.Message.Reference(),
			Files:     files,
		})
		return err
	}
	if ctx.deferred {
		_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Files: files,
		})
		return err
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: files,
		},
	})
}

// FindImageURL attempts to locate an image URL from the context.
func (ctx *OperationContext) FindImageURL() (string, error) {
	if ctx.Message != nil {
		return FindImageURLFromMessage(ctx.Message)
	}
	return FindImageURLInChannel(ctx.Session, ctx.Interaction.ChannelID, "")
}

// FindVideoURL attempts to locate a video URL from the context.
func (ctx *OperationContext) FindVideoURL() (string, error) {
	if ctx.Message != nil {
		return FindVideoURLFromMessage(ctx.Message)
	}
	return FindVideoURLInChannel(ctx.Session, ctx.Interaction.ChannelID, "")
}

func TypingIndicatorForContext(ctx *OperationContext) func() {
	if ctx.Message != nil {
		return TypingIndicator(ctx.Message)
	}
	return func() {}
}

// TypingIndicator invokes a typing indicator in the channel of a message
func TypingIndicator(message *discordgo.MessageCreate) func() {
	stopTyping := Schedule(
		func() {
			log.Debug().Str("channel", message.ChannelID).Msg("Invoking typing indicator in channel")
			err := Instance.session.ChannelTyping(message.ChannelID)
			if err != nil {
				log.Error().Err(err).Msg("Error while attempting invoke typing indicator in channel")
				return
			}
		},
		5*time.Second,
	)
	return func() {
		stopTyping <- true
	}
}

// Schedule some func to be run in a cancelable goroutine on an interval
func Schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// ImageURLFromComponent attempts to retrieve an image URL from a given component, recursing into child components if necessary
func ImageURLFromComponent(component discordgo.MessageComponent) string {
	switch c := component.(type) {
	case *discordgo.MediaGallery:
		if len(c.Items) == 0 {
			return ""
		}

		return c.Items[0].Media.URL
	case *discordgo.Container:
		for _, child := range c.Components {
			url := ImageURLFromComponent(child)
			if url != "" {
				return url
			}
		}

		return ""
	default:
		return ""
	}
}

// ImageURLFromMessage attempts to retrieve an image URL from a given message.
func ImageURLFromMessage(m *discordgo.Message) string {
	for _, embed := range m.Embeds {
		if embed.Type == "Image" {
			return embed.URL
		} else if embed.Image != nil {
			return embed.Image.URL
		}
	}

	for _, attachment := range m.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image/") {
			return attachment.URL
		}
	}

	for _, component := range m.Components {
		url := ImageURLFromComponent(component)
		if url != "" {
			return url
		}
	}

	return ""
}

// VideoURLFromMessage attempts to retrieve a video URL from a given message.
func VideoURLFromMessage(m *discordgo.Message) string {
	for _, embed := range m.Embeds {
		if embed.Video != nil && embed.Video.URL != "" {
			return embed.Video.URL
		}
		if (embed.Type == discordgo.EmbedTypeVideo || embed.Type == discordgo.EmbedTypeGifv) && embed.URL != "" {
			return embed.URL
		}
	}

	for _, attachment := range m.Attachments {
		if IsVideoAttachment(attachment) {
			return attachment.URL
		}
	}

	for _, component := range m.Components {
		url := ImageURLFromComponent(component)
		if url != "" {
			return url
		}
	}

	return ""
}

func IsVideoAttachment(attachment *discordgo.MessageAttachment) bool {
	if strings.HasPrefix(attachment.ContentType, "video/") {
		return true
	}

	switch strings.ToLower(path.Ext(attachment.Filename)) {
	case ".avi", ".m4v", ".mkv", ".mov", ".mp4", ".webm":
		return true
	default:
		return false
	}
}

// FindImageURLFromMessage attempts to find an image in a given message, falling back to scanning message history if one cannot be found.
func FindImageURLFromMessage(m *discordgo.MessageCreate) (string, error) {
	if imageUrl := ImageURLFromMessage(m.Message); imageUrl != "" {
		return imageUrl, nil
	}

	if m.ReferencedMessage != nil {
		if imageUrl := ImageURLFromMessage(m.ReferencedMessage); imageUrl != "" {
			return imageUrl, nil
		}
	}

	return FindImageURLInChannel(Instance.session, m.ChannelID, m.ID)
}

// FindVideoURLFromMessage attempts to find a video in a given message, falling back to scanning message history if one cannot be found.
func FindVideoURLFromMessage(m *discordgo.MessageCreate) (string, error) {
	if videoUrl := VideoURLFromMessage(m.Message); videoUrl != "" {
		return videoUrl, nil
	}

	if m.ReferencedMessage != nil {
		if videoUrl := VideoURLFromMessage(m.ReferencedMessage); videoUrl != "" {
			return videoUrl, nil
		}
	}

	return FindVideoURLInChannel(Instance.session, m.ChannelID, m.ID)
}

func FindImageURLInChannel(s *discordgo.Session, channelID string, beforeID string) (string, error) {
	messages, err := s.ChannelMessages(channelID, 20, beforeID, "", "")
	if err != nil {
		return "", fmt.Errorf("error retrieving message history: %w", err)
	}

	for _, message := range messages {
		if imageUrl := ImageURLFromMessage(message); imageUrl != "" {
			return imageUrl, nil
		}
	}
	return "", errors.New("unable to locate an image")
}

func FindVideoURLInChannel(s *discordgo.Session, channelID string, beforeID string) (string, error) {
	messages, err := s.ChannelMessages(channelID, 20, beforeID, "", "")
	if err != nil {
		return "", fmt.Errorf("error retrieving message history: %w", err)
	}

	for _, message := range messages {
		if videoUrl := VideoURLFromMessage(message); videoUrl != "" {
			return videoUrl, nil
		}
	}
	return "", errors.New("unable to locate a video")
}
