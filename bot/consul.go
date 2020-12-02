package bot

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/consul/api"
)

// ConsulBackend represents a persistence backend that persists data via Consul K/V.
type ConsulBackend struct {
	consulClient *api.Client
}

// Get retrieves a value from the Consul backend.
func (backend *ConsulBackend) Get(key string, valuePtr interface{}) error {
	key = "borik/" + key
	kvPair, _, err := backend.consulClient.KV().Get(key, nil)
	if err != nil {
		return fmt.Errorf("error retrieving value from consul backend: %w", err)
	}

	if kvPair == nil {
		return nil
	}

	err = json.Unmarshal(kvPair.Value, valuePtr)
	if err != nil {
		return fmt.Errorf("error parsing returned json: %w", err)
	}

	return nil
}

// Put stores a value in the Consul backend.
func (backend *ConsulBackend) Put(key string, value interface{}) error {
	key = "borik/" + key
	val, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error dumping json data: %w", err)
	}

	_, err = backend.consulClient.KV().Put(&api.KVPair{Key: key, Value: val}, nil)
	if err != nil {
		return fmt.Errorf("error writing to consul: %w", err)
	}

	return nil
}

var _ PersistenceBackend = (*ConsulBackend)(nil)

// NewConsulBackend constructs a new Consul-backed storage backend.
func NewConsulBackend(config Config) (*ConsulBackend, error) {
	client, err := api.NewClient(&api.Config{Address: config.ConsulAddress})
	if err != nil {
		return &ConsulBackend{}, fmt.Errorf("error creating consul client: %w", err)
	}
	_, err = client.Status().Leader()
	if err != nil {
		return &ConsulBackend{}, fmt.Errorf("error checking consul connection status: %w", err)
	}
	backend := ConsulBackend{client}
	return &backend, nil
}
