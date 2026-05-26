package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/delivery"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func startTypingPulse(ctx workflow.Context, target delivery.Target) workflow.CancelFunc {
	if target.Type != delivery.TargetTypeMessage || target.ChannelID == "" {
		return func() {}
	}

	typingCtx, cancel := workflow.WithCancel(ctx)
	workflow.Go(typingCtx, func(ctx workflow.Context) {
		for {
			if err := workflow.ExecuteActivity(ctx, activities.SendDiscordTypingActivityName, target).Get(ctx, nil); err != nil {
				workflow.GetLogger(ctx).Warn("Error sending typing indicator", "error", err)
			}

			if err := workflow.NewTimer(ctx, 5*time.Second).Get(ctx, nil); err != nil {
				return
			}
		}
	})

	return cancel
}

func notifyFailure(ctx workflow.Context, target delivery.Target, err error) {
	if target.IsZero() {
		return
	}

	future := workflow.ExecuteActivity(ctx, activities.SendDiscordFailureActivityName, target, err.Error())
	if notifyErr := future.Get(ctx, nil); notifyErr != nil {
		workflow.GetLogger(ctx).Warn("Error sending failure message", "error", notifyErr)
	}
}

func cleanupWorkspace(ctx workflow.Context, jobWorkspace workspace.Workspace) {
	future := workflow.ExecuteActivity(ctx, activities.CleanupWorkspace, jobWorkspace)
	if cleanupErr := future.Get(ctx, nil); cleanupErr != nil {
		workflow.GetLogger(ctx).Warn("Error cleaning up workspace", "error", cleanupErr)
	}
}

func sendResult(
	ctx workflow.Context,
	jobWorkspace workspace.Workspace,
	artifact workspace.Artifact,
	target delivery.Target,
) error {
	if target.IsZero() {
		return nil
	}

	future := workflow.ExecuteActivity(ctx, activities.SendDiscordResultActivityName, jobWorkspace, artifact, target)
	if err := future.Get(ctx, nil); err != nil {
		return fmt.Errorf("error sending result: %w", err)
	}
	return nil
}
