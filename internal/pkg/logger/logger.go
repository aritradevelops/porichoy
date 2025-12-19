package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var instance zerolog.Logger

func init() {
	if os.Getenv("ENV") == "production" {
		instance = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		instance = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	}
}

func Debug() *zerolog.Event {
	return instance.Debug()
}

func Info() *zerolog.Event {
	return instance.Info()
}

func Warn() *zerolog.Event {
	return instance.Warn()
}

func Error() *zerolog.Event {
	return instance.Error()
}

func Fatal() *zerolog.Event {
	return instance.Fatal()
}

func Panic() *zerolog.Event {
	return instance.Panic()
}
