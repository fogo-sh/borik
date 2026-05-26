package worker

import (
	"fmt"
	"math"
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/sysinfo"
	"go.temporal.io/sdk/worker"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/workflows"
	"github.com/fogo-sh/borik/pkg/logging"
)

type Worker struct {
	client         client.Client
	worker         worker.Worker
	discordWorker  worker.Worker
	discordSession *discordgo.Session
	interruptChan  chan any
}

func (w *Worker) Start() error {
	defer w.client.Close()

	if err := w.worker.Start(); err != nil {
		return fmt.Errorf("error starting temporal worker: %w", err)
	}
	defer w.worker.Stop()

	if err := w.discordWorker.Start(); err != nil {
		return fmt.Errorf("error starting temporal discord worker: %w", err)
	}
	defer w.discordWorker.Stop()

	<-w.interruptChan

	return nil
}

func (w *Worker) Stop() {
	w.interruptChan <- struct{}{}
}

func cgroupAwareCoreCount() int {
	cores := runtime.GOMAXPROCS(0)
	if cores < 1 {
		return 1
	}

	return cores
}

func New() (*Worker, error) {
	token := strings.TrimSpace(config.Instance.Token)
	if token == "" {
		return nil, fmt.Errorf("discord bot token must be set")
	}

	discordSession, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating discord session: %w", err)
	}

	c, err := client.Dial(client.Options{
		Logger:    logging.NewTemporalLogger(),
		Namespace: config.Instance.TemporalNamespace,
		HostPort:  config.Instance.TemporalHostPort,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating temporal client: %w", err)
	}

	w := worker.New(
		c,
		config.Instance.TemporalQueueName,
		worker.Options{
			SysInfoProvider:                    sysinfo.SysInfoProvider(),
			MaxConcurrentActivityExecutionSize: cgroupAwareCoreCount(),
		},
	)
	workflows.RegisterWorkflows(w)
	activities.RegisterActivities(w)

	discordWorker := worker.New(
		c,
		config.Instance.TemporalDiscordQueueName,
		worker.Options{
			SysInfoProvider:                    sysinfo.SysInfoProvider(),
			MaxConcurrentActivityExecutionSize: int(math.Max(float64(cgroupAwareCoreCount()*8), 16)),
		},
	)
	activities.RegisterDiscordActivities(discordWorker, discordSession)

	return &Worker{
		client:         c,
		worker:         w,
		discordWorker:  discordWorker,
		discordSession: discordSession,
		interruptChan:  make(chan any),
	}, nil
}
