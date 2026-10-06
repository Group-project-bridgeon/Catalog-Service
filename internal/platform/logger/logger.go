package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func New(service string) zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Str("service", service).
		Logger()
}
