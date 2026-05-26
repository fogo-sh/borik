package activities

import (
	"context"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Waaw(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyMirror(jobWorkspace, opArgs, mirrorDirectionHorizontal, true)
}

func Haah(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyMirror(jobWorkspace, opArgs, mirrorDirectionHorizontal, false)
}

func Woow(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyMirror(jobWorkspace, opArgs, mirrorDirectionVertical, false)
}

func Hooh(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyMirror(jobWorkspace, opArgs, mirrorDirectionVertical, true)
}
