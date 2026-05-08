package activities

import (
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

type OperationArgs struct {
	Frame workspace.Artifact
	Args  any
}
