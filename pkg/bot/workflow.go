package bot

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.temporal.io/sdk/client"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/workflows"
)

type testArgs struct {
	ImageURL string `description:"Image URL to test"`
}

func (b *Bot) workflowTestCommand(message *discordgo.MessageCreate, args testArgs) {
	we, err := b.temporalClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        "test-workflow",
			TaskQueue: config.Instance.TemporalQueueName,
		},
		workflows.TestWorkflow,
		args.ImageURL,
	)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error executing workflow: "+err.Error())
		return
	}

	var result []byte
	err = we.Get(context.Background(), &result)
	if err != nil {
		b.session.ChannelMessageSend(message.ChannelID, "Error getting workflow result: "+err.Error())
		return
	}

	b.session.ChannelMessageSend(message.ChannelID, "Workflow result: "+string(result))
}
