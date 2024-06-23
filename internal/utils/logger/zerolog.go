package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	l   *zerolog.Logger
	app string
}

func NewLogger(env string, app string) Logger {

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	level := zerolog.DebugLevel
	// if env == "prod" {
	level = zerolog.WarnLevel
	// }
	logger = logger.Level(level)
	return Logger{l: &logger, app: app}
}

func (l Logger) Infof(ctx context.Context, message string, args ...interface{}) {
	l.l.Info().Msgf("%s:  %s", l.app, fmt.Sprintf(message, args...))
}

func (l Logger) Debugf(ctx context.Context, message string, args ...interface{}) {
	l.l.Debug().Msgf("%s:  %s", l.app, fmt.Sprintf(message, args...))
}

func (l Logger) ErrorMessage(ctx context.Context, message string, args ...interface{}) {
	l.l.Error().Msgf("%s:  %s", l.app, fmt.Sprintf(message, args...))
}

func (l Logger) Warnf(ctx context.Context, message string, args ...interface{}) {
	l.l.Warn().Msgf("%s:  %s", l.app, fmt.Sprintf(message, args...))
}

func (l Logger) Errorf(ctx context.Context, err error, message string, args ...interface{}) {
	l.l.Error().Err(err).Msgf("%s:  %s", l.app, fmt.Sprintf(message, args...))
}

func (l Logger) Fatalf(ctx context.Context, message string, args ...interface{}) {
	l.l.Fatal().Msgf("%s:  %s", l.app, fmt.Sprintf(message, args...))
}
