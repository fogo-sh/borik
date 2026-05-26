package activities

import (
	"context"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func InitJobWorkspace(ctx context.Context, jobID string) (workspace.Workspace, error) {
	return workspace.InitJobWorkspace(jobID)
}
