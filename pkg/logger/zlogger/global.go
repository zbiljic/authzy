package zlogger

import (
	"context"
	"fmt"
	"sync"

	"github.com/kelseyhightower/envconfig"

	"github.com/zbiljic/authzy/pkg/logger"
)

var (
	global logger.Logger
	mu     sync.Mutex
)

// Initialize this logger once
var once sync.Once

func init() {
	once.Do(initLogger)
}

func initLogger() {
	config := logger.Config{}

	err := envconfig.Process("", &config)
	if err != nil {
		panic(err)
	}

	err = SetupLogging(&config)
	if err != nil {
		panic(err)
	}
}

// SetupLogging configures the global logger.
func SetupLogging(config *logger.Config) error {
	mu.Lock()
	defer mu.Unlock()

	var err error

	global, err = New(config)
	if err != nil {
		return fmt.Errorf("error creating global logger: %w", err)
	}

	return nil
}

func Instance() logger.Logger {
	return global
}

func Debug(args ...interface{}) {
	global.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	global.Debugf(format, args...)
}

func Info(args ...interface{}) {
	global.Info(args...)
}

func Infof(format string, args ...interface{}) {
	global.Infof(format, args...)
}

func Warn(args ...interface{}) {
	global.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	global.Warnf(format, args...)
}

func Error(args ...interface{}) {
	global.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	global.Errorf(format, args...)
}

func WithFields(fields logger.Fields) logger.Logger {
	return global.WithFields(fields)
}

func NewContext(ctx context.Context, fields logger.Fields) context.Context {
	return global.NewContext(ctx, fields)
}

func WithContext(ctx context.Context) logger.Logger {
	return global.WithContext(ctx)
}
