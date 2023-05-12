package worker

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/logging"
)

type Worker struct {
	client        client.Client
	worker        worker.Worker
	interruptChan chan interface{}
}

func (w *Worker) Start() error {
	defer w.client.Close()

	err := w.worker.Run(w.interruptChan)
	if err != nil {
		return fmt.Errorf("error starting temporal worker: %w", err)
	}

	return nil
}

func (w *Worker) Stop() {
	w.interruptChan <- struct{}{}
}

func New() (*Worker, error) {
	c, err := client.Dial(client.Options{
		Logger:    logging.NewTemporalLogger(),
		Namespace: config.Instance.TemporalNamespace,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating temporal client: %w", err)
	}

	w := worker.New(c, config.Instance.TemporalQueueName, worker.Options{})

	return &Worker{
		client:        c,
		worker:        w,
		interruptChan: make(chan interface{}),
	}, nil
}
