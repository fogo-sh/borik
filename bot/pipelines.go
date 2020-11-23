package bot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ErrNoPendingPipeline occurs when attempting to add a step to a command pipeline for a user with no pending pipeline.
var ErrNoPendingPipeline = errors.New("you do not have a pending pipeline - use the createpipeline command to begin creating one")

// ErrPendingPipelineExists occurs when attempting to create a new pipeline for a user with an already existing pending pipeline.
var ErrPendingPipelineExists = errors.New("you already have an existing pending pipeline")

// PipelineEntry represents a single entry in a command pipeline.
type PipelineEntry struct {
	Operation string
	Args      interface{}
}

// PipelineManager manages the saving, creation, and execution of command pipelines.
type PipelineManager struct {
	PendingPipelines map[string][]PipelineEntry
}

// CreatePipeline creates a pending pipeline for a given user.
func (manager *PipelineManager) CreatePipeline(owner string) error {
	_, found := manager.PendingPipelines[owner]
	if found {
		return ErrPendingPipelineExists
	}
	manager.PendingPipelines[owner] = make([]PipelineEntry, 0)
	return nil
}

// AddStep adds a step to a user's pending command pipeline.
func (manager *PipelineManager) AddStep(message *discordgo.MessageCreate, operation string, args interface{}) error {
	pipeline, ok := manager.PendingPipelines[message.Author.ID]
	if !ok {
		return ErrNoPendingPipeline
	}

	manager.PendingPipelines[message.Author.ID] = append(pipeline, PipelineEntry{operation, args})

	return nil
}

func _CreatePipelineCommand(message *discordgo.MessageCreate, args struct{}) {
	err := Instance.PipelineManager.CreatePipeline(message.Author.ID)
	if err != nil {
		Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror creating new pipeline: %s\n```", err.Error()))
	}
}
