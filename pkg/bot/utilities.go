package bot

import (
	"fmt"
	"io"
	"mime"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var messageURLRegex = regexp.MustCompile(`(?i)https?://[^\s<>"']+`)

type mediaType struct {
	name         string
	urlFromEmbed func(*discordgo.MessageEmbed) string
}

var (
	imageMediaType = mediaType{name: "image", urlFromEmbed: imageURLFromEmbed}
	videoMediaType = mediaType{name: "video", urlFromEmbed: videoURLFromEmbed}
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
	return ctx.findMediaURL(imageMediaType)
}

// FindVideoURL attempts to locate a video URL from the context.
func (ctx *OperationContext) FindVideoURL() (string, error) {
	return ctx.findMediaURL(videoMediaType)
}

func (ctx *OperationContext) findMediaURL(kind mediaType) (string, error) {
	if ctx.Message != nil {
		return findMediaURLFromMessage(ctx.Message, kind)
	}
	return findMediaURLInChannel(ctx.Session, ctx.Interaction.ChannelID, "", kind)
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

func mediaURLFromComponent(component discordgo.MessageComponent) string {
	switch c := component.(type) {
	case *discordgo.MediaGallery:
		if len(c.Items) == 0 {
			return ""
		}

		return c.Items[0].Media.URL
	case *discordgo.Container:
		for _, child := range c.Components {
			url := mediaURLFromComponent(child)
			if url != "" {
				return url
			}
		}

		return ""
	default:
		return ""
	}
}

func mediaURLFromContent(content string, kind mediaType) string {
	contentTypePrefix := kind.name + "/"

	for _, candidate := range messageURLRegex.FindAllString(content, -1) {
		candidate = strings.TrimRight(candidate, ".,!?;:)]}")
		parsedURL, err := url.Parse(candidate)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			continue
		}

		contentType := mime.TypeByExtension(strings.ToLower(path.Ext(parsedURL.Path)))
		if strings.HasPrefix(contentType, contentTypePrefix) {
			return candidate
		}
	}

	return ""
}

func mediaURLFromMessage(m *discordgo.Message, kind mediaType) string {
	for _, embed := range m.Embeds {
		if url := kind.urlFromEmbed(embed); url != "" {
			return url
		}
	}

	for _, attachment := range m.Attachments {
		if attachmentMatchesMediaType(attachment, kind) {
			return attachment.URL
		}
	}

	for _, component := range m.Components {
		if url := mediaURLFromComponent(component); url != "" {
			return url
		}
	}

	if url := mediaURLFromContent(m.Content, kind); url != "" {
		return url
	}

	return ""
}

func imageURLFromEmbed(embed *discordgo.MessageEmbed) string {
	if embed.Type == discordgo.EmbedTypeImage && embed.URL != "" {
		return embed.URL
	}
	if embed.Image != nil {
		return embed.Image.URL
	}

	return ""
}

func videoURLFromEmbed(embed *discordgo.MessageEmbed) string {
	if embed.Video != nil && embed.Video.URL != "" {
		return embed.Video.URL
	}
	if (embed.Type == discordgo.EmbedTypeVideo || embed.Type == discordgo.EmbedTypeGifv) && embed.URL != "" {
		return embed.URL
	}

	return ""
}

func attachmentMatchesMediaType(attachment *discordgo.MessageAttachment, kind mediaType) bool {
	contentTypePrefix := kind.name + "/"

	if strings.HasPrefix(attachment.ContentType, contentTypePrefix) {
		return true
	}

	contentType := mime.TypeByExtension(strings.ToLower(path.Ext(attachment.Filename)))
	return strings.HasPrefix(contentType, contentTypePrefix)
}

func findMediaURLFromMessage(m *discordgo.MessageCreate, kind mediaType) (string, error) {
	if mediaURL := mediaURLFromMessage(m.Message, kind); mediaURL != "" {
		return mediaURL, nil
	}

	if m.ReferencedMessage != nil {
		if mediaURL := mediaURLFromMessage(m.ReferencedMessage, kind); mediaURL != "" {
			return mediaURL, nil
		}
	}

	return findMediaURLInChannel(Instance.session, m.ChannelID, m.ID, kind)
}

func findMediaURLInChannel(s *discordgo.Session, channelID string, beforeID string, kind mediaType) (string, error) {
	messages, err := s.ChannelMessages(channelID, 20, beforeID, "", "")
	if err != nil {
		return "", fmt.Errorf("error retrieving message history: %w", err)
	}

	for _, message := range messages {
		if mediaURL := mediaURLFromMessage(message, kind); mediaURL != "" {
			return mediaURL, nil
		}
	}
	return "", fmt.Errorf("unable to locate a %s", kind.name)
}

func closeBody(body io.Closer, message string) {
	if err := body.Close(); err != nil {
		log.Error().Err(err).Msg(message)
	}
}
