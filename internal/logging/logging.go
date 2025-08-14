package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New() zerolog.Logger {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = time.RFC3339
	return l
}
