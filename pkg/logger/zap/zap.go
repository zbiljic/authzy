package zap

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zbiljic/authzy/pkg/logger"
)

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

func toZapLevel(level string) zapcore.Level {
	switch level {
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.DebugLevel:
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}

func getEncoder(logFormat string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time" // This will change the key from 'ts' to 'time'
	// RFC3339-formatted string for time
	encoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(logger.RFC3339Milli))
	})

	if logger.JSONFormat == logFormat {
		return zapcore.NewJSONEncoder(encoderConfig)
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func NewLogger(logLevel, logFormat string) (logger.Logger, error) {
	cores := []zapcore.Core{}

	level := toZapLevel(logLevel)
	encoder := getEncoder(logFormat)
	writer := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(encoder, writer, level)
	cores = append(cores, core)

	combinedCore := zapcore.NewTee(cores...)

	// AddCallerSkip skips 2 number of callers, this is important else the file
	// that gets logged will always be the wrapped file. In our case zap.go
	zLogger := zap.New(combinedCore,
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	).Sugar()

	l := &zapLogger{
		sugaredLogger: zLogger,
	}

	return l, nil
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.sugaredLogger.Debug(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Info(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.sugaredLogger.Warn(args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Error(args ...interface{}) {
	l.sugaredLogger.Error(args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) WithFields(fields logger.Fields) logger.Logger {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}

	newLogger := l.sugaredLogger.With(f...)

	return &zapLogger{newLogger}
}

// fieldsContextKey is used to reference logger fields as context value.
type fieldsContextKey struct{}

func (l *zapLogger) NewContext(ctx context.Context, fields logger.Fields) context.Context {
	if ctxFields, ok := ctx.Value(fieldsContextKey{}).(logger.Fields); ok {
		// extend context fields
		for k, v := range fields {
			ctxFields[k] = v
		}
		return context.WithValue(ctx, fieldsContextKey{}, ctxFields)
	}
	return context.WithValue(ctx, fieldsContextKey{}, fields)
}

func (l *zapLogger) WithContext(ctx context.Context) logger.Logger {
	if ctxFields, ok := ctx.Value(fieldsContextKey{}).(logger.Fields); ok {
		return l.WithFields(ctxFields)
	}
	return l
}
