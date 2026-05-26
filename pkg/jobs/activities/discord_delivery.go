package activities

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/fogo-sh/borik/pkg/jobs/delivery"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

const (
	SendDiscordResultActivityName  = "SendDiscordResult"
	SendDiscordFailureActivityName = "SendDiscordFailure"
	SendDiscordTypingActivityName  = "SendDiscordTyping"
)

type DiscordDeliveryActivities struct {
	Session *discordgo.Session
}

func (a DiscordDeliveryActivities) SendDiscordTyping(ctx context.Context, target delivery.Target) error {
	if target.Type != delivery.TargetTypeMessage || target.ChannelID == "" {
		return nil
	}

	return a.Session.ChannelTyping(target.ChannelID)
}

func (a DiscordDeliveryActivities) SendDiscordFailure(
	ctx context.Context,
	target delivery.Target,
	message string,
) error {
	if target.IsZero() {
		return nil
	}

	content := message
	if target.FailureMessagePrefix != "" {
		content = fmt.Sprintf("%s: %s", target.FailureMessagePrefix, message)
	}

	switch target.Type {
	case delivery.TargetTypeNone:
		return nil
	case delivery.TargetTypeMessage:
		_, err := a.Session.ChannelMessageSendReply(target.ChannelID, content, messageReference(target))
		if err != nil {
			return fmt.Errorf("error sending discord failure message: %w", err)
		}
	case delivery.TargetTypeInteraction:
		_, err := a.Session.InteractionResponseEdit(interaction(target), &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			return fmt.Errorf("error sending discord failure message: %w", err)
		}
	default:
		return fmt.Errorf("unknown discord delivery target type %q", target.Type)
	}
	return nil
}

func (a DiscordDeliveryActivities) SendDiscordResult(
	ctx context.Context,
	jobWorkspace workspace.Workspace,
	artifact workspace.Artifact,
	target delivery.Target,
) error {
	if target.IsZero() {
		return nil
	}

	image, err := jobWorkspace.Retrieve(artifact)
	if err != nil {
		return fmt.Errorf("error retrieving result artifact: %w", err)
	}

	defer func() {
		if err := jobWorkspace.Cleanup(); err != nil {
			log.Error().Err(err).Msg("Error cleaning up workspace")
		}
	}()

	file := &discordgo.File{
		Name:        target.Filename,
		ContentType: target.ContentType,
		Reader:      bytes.NewReader(image),
	}

	switch target.Type {
	case delivery.TargetTypeNone:
		return nil
	case delivery.TargetTypeMessage:
		_, err := a.Session.ChannelMessageSendComplex(target.ChannelID, &discordgo.MessageSend{
			Reference: messageReference(target),
			Files:     []*discordgo.File{file},
		})
		if err != nil {
			return fmt.Errorf("error sending discord result: %w", err)
		}
	case delivery.TargetTypeInteraction:
		_, err := a.Session.InteractionResponseEdit(interaction(target), &discordgo.WebhookEdit{
			Files: []*discordgo.File{file},
		})
		if err != nil {
			return fmt.Errorf("error sending discord result: %w", err)
		}
	default:
		return fmt.Errorf("unknown discord delivery target type %q", target.Type)
	}
	return nil
}

func CleanupWorkspace(ctx context.Context, jobWorkspace workspace.Workspace) error {
	return jobWorkspace.Cleanup()
}

func messageReference(target delivery.Target) *discordgo.MessageReference {
	if target.MessageID == "" {
		return nil
	}

	failIfNotExists := false
	return &discordgo.MessageReference{
		Type:            discordgo.MessageReferenceTypeDefault,
		MessageID:       target.MessageID,
		ChannelID:       target.ChannelID,
		GuildID:         target.GuildID,
		FailIfNotExists: &failIfNotExists,
	}
}

func interaction(target delivery.Target) *discordgo.Interaction {
	return &discordgo.Interaction{
		ID:      target.InteractionID,
		AppID:   target.AppID,
		Token:   target.InteractionToken,
		GuildID: target.GuildID,
	}
}
