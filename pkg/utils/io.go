package utils

import (
	"io"

	"github.com/rs/zerolog/log"
)

func CloseBody(body io.Closer, message string) {
	if err := body.Close(); err != nil {
		log.Error().Err(err).Msg(message)
	}
}
