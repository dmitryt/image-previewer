package logger

import (
	"errors"
	"os"

	"github.com/dmitryt/image-previewer/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var ErrFileLog = errors.New("cannot setup file log")

func getLogLevel(str string) zerolog.Level {
	switch str {
	case "error":
		return zerolog.ErrorLevel
	case "warn":
		return zerolog.WarnLevel
	case "info":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	default:
		return zerolog.InfoLevel
	}
}

func Init(c *config.Config) {
	zerolog.SetGlobalLevel(getLogLevel(c.LogLevel))
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})
}
