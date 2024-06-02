package workspace

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/google/uuid"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/config"
)

func EnsureExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return fmt.Errorf("error creating workspace: %w", err)
		}
	}
	return nil
}

type Artifact string

type Workspace struct {
	Path string
}

func (w Workspace) Persist(data []byte) (Artifact, error) {
	artifactIdentifier := uuid.New().String()

	artifactPath := path.Join(w.Path, artifactIdentifier)

	f, err := os.OpenFile(artifactPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("error creating artifact file: %w", err)
	}

	_, err = f.Write(data)
	if err != nil {
		return "", fmt.Errorf("error writing artifact data: %w", err)
	}

	return Artifact(artifactIdentifier), nil
}

func (w Workspace) PersistWand(wand *imagick.MagickWand) (Artifact, error) {
	data := wand.GetImageBlob()
	return w.Persist(data)
}

func (w Workspace) Retrieve(artifact Artifact) ([]byte, error) {
	artifactPath := path.Join(w.Path, string(artifact))

	f, err := os.Open(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("error opening artifact file: %w", err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading artifact data: %w", err)
	}

	return data, nil
}

func (w Workspace) RetrieveWand(artifact Artifact) (*imagick.MagickWand, error) {
	data, err := w.Retrieve(artifact)
	if err != nil {
		return nil, err
	}

	wand := imagick.NewMagickWand()
	err = wand.ReadImageBlob(data)
	if err != nil {
		return nil, fmt.Errorf("error reading image blob: %w", err)
	}

	return wand, nil
}

func (w Workspace) Cleanup() error {
	return os.RemoveAll(w.Path)
}

func InitJobWorkspace(jobID string) (Workspace, error) {
	err := EnsureExists(config.Instance.WorkspacePath)
	if err != nil {
		return Workspace{}, fmt.Errorf("error creating workspace: %w", err)
	}

	jobPath := path.Join(config.Instance.WorkspacePath, jobID)

	err = EnsureExists(jobPath)
	if err != nil {
		return Workspace{}, fmt.Errorf("error creating job workspace: %w", err)
	}

	return Workspace{Path: jobPath}, nil
}
