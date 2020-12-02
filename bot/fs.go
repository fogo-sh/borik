package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FSBackend represents a persistence backend that persists data via the filesystem.
type FSBackend struct {
	backendPath string
}

// Get retrieves a value from the filesystem backend.
func (backend *FSBackend) Get(key string, valuePtr interface{}) error {
	objFile, err := os.Open(filepath.Join(backend.backendPath, key))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error attempting to open key file: %w", err)
		}
		return nil
	}
	defer objFile.Close()

	objBytes, _ := ioutil.ReadAll(objFile)

	err = json.Unmarshal(objBytes, valuePtr)
	if err != nil {
		return fmt.Errorf("error parsing json: %w", err)
	}

	return nil
}

// Put stores a value in the filesystem backend.
func (backend *FSBackend) Put(key string, value interface{}) error {
	fspath := filepath.Join(backend.backendPath, key)
	dir := filepath.Dir(fspath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating key directory: %w", err)
	}

	objBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshalling json: %w", err)
	}

	err = ioutil.WriteFile(fspath, objBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing key to disk: %w", err)
	}

	return nil
}

var _ PersistenceBackend = (*FSBackend)(nil)

// NewFSBackend constructs a new Filesystem-backed storage backend.
func NewFSBackend(config Config) (*FSBackend, error) {
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		err := os.MkdirAll(config.FilePath, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating backend directory: %w", err)
		}
	}
	backend := FSBackend{config.FilePath}
	return &backend, nil
}
