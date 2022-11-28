package logger

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	logger *zerolog.Logger
	once   sync.Once
)

// GetLogger is getter to create new logger instance.
func GetLogger() *zerolog.Logger {
	once.Do(func() {
		logger = newLogger()
	})

	return logger
}

func newLogger() *zerolog.Logger {
	zerolog.ErrorFieldName = "message"

	return &log.Logger
}
