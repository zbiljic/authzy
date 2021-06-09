package zerolog

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"

	"github.com/zbiljic/authzy/pkg/logger"
)

var (
	// noColor defines if the output is colorized or not. It's dynamically set to
	// false or true based on the stdout's file descriptor referring to a terminal
	// or not.
	noColor = os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))
)

type zerologLogger struct {
	logger *zerolog.Logger
}

func toZerologLevel(level string) zerolog.Level {
	switch level {
	case logger.ErrorLevel:
		return zerolog.ErrorLevel
	case logger.WarnLevel:
		return zerolog.WarnLevel
	case logger.InfoLevel:
		return zerolog.InfoLevel
	case logger.DebugLevel:
		return zerolog.DebugLevel
	default:
		return zerolog.InfoLevel
	}
}

func NewLogger(logLevel, logFormat string) (logger.Logger, error) {
	level := toZerologLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	var zLogger zerolog.Logger

	if logger.JSONFormat == logFormat {
		zerolog.TimeFieldFormat = logger.RFC3339Milli
		zLogger = zerolog.New(os.Stderr)
	} else {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    noColor,
			TimeFormat: time.RFC3339,
		}
		zLogger = zerolog.New(output)
	}

	zLogger = zLogger.With().Timestamp().Logger()

	l := &zerologLogger{
		logger: &zLogger,
	}

	return l, nil
}

func (l *zerologLogger) Debug(args ...interface{}) {
	l.logger.Debug().Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *zerologLogger) Info(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *zerologLogger) Warn(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *zerologLogger) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *zerologLogger) WithFields(fields logger.Fields) logger.Logger {
	logCtx := l.logger.With()
	for k, v := range fields {
		logCtx = logCtx.Str(k, fmt.Sprint(v))
	}

	newLogger := logCtx.Logger()

	return &zerologLogger{&newLogger}
}

// fieldsContextKey is used to reference logger fields as context value.
type fieldsContextKey struct{}

func (l *zerologLogger) NewContext(ctx context.Context, fields logger.Fields) context.Context {
	if ctxFields, ok := ctx.Value(fieldsContextKey{}).(logger.Fields); ok {
		// extend context fields
		for k, v := range fields {
			ctxFields[k] = v
		}
		return context.WithValue(ctx, fieldsContextKey{}, ctxFields)
	}
	return context.WithValue(ctx, fieldsContextKey{}, fields)
}

func (l *zerologLogger) WithContext(ctx context.Context) logger.Logger {
	if ctxFields, ok := ctx.Value(fieldsContextKey{}).(logger.Fields); ok {
		return l.WithFields(ctxFields)
	}
	return l
}
