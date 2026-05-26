package logging

import (
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
	"go.temporal.io/sdk/log"
)

type TemporalZerologShim struct {
	logger zerolog.Logger
}

func (t *TemporalZerologShim) Debug(msg string, keyvals ...any) {
	t.logger.Debug().Fields(keyvals).Msg(msg)
}

func (t *TemporalZerologShim) Info(msg string, keyvals ...any) {
	t.logger.Info().Fields(keyvals).Msg(msg)
}

func (t *TemporalZerologShim) Warn(msg string, keyvals ...any) {
	t.logger.Warn().Fields(keyvals).Msg(msg)
}

func (t *TemporalZerologShim) Error(msg string, keyvals ...any) {
	t.logger.Error().Fields(keyvals).Msg(msg)
}

var _ log.Logger = (*TemporalZerologShim)(nil)

func NewTemporalLogger() log.Logger {
	return &TemporalZerologShim{
		logger: zl.Logger,
	}
}
