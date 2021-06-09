package logger

import "context"

const (
	RFC3339Milli = "2006-01-02T15:04:05.999Z07:00" // time.RFC3339 with millisecond precision
)

const (
	// DebugLevel has verbose message.
	DebugLevel = "debug"
	// InfoLevel is default log level.
	InfoLevel = "info"
	// WarnLevel is for logging messages about possible issues.
	WarnLevel = "warn"
	// ErrorLevel is for logging errors.
	ErrorLevel = "error"
)

// The log format can either be text or JSON.
const (
	JSONFormat = "json"
	TextFormat = "text"
)

// Logger defines an interface for logging.
type Logger interface {
	// Debug logs a debug message, patterned after log.Print.
	Debug(args ...interface{})
	// Debugf logs a debug message, patterned after log.Printf.
	Debugf(format string, args ...interface{})

	// Info logs an information message, patterned after log.Print.
	Info(args ...interface{})
	// Infof logs an information message, patterned after log.Printf.
	Infof(format string, args ...interface{})

	// Warn logs a warning message, patterned after log.Print.
	Warn(args ...interface{})
	// Warnf logs a warning message, patterned after log.Printf.
	Warnf(format string, args ...interface{})

	// Error logs an error message, patterned after log.Print.
	Error(args ...interface{})
	// Errorf logs an error message, patterned after log.Printf.
	Errorf(format string, args ...interface{})

	// WithFields adds a struct of fields to the log entry.
	WithFields(fields Fields) Logger

	// NewContext adds a struct of fields to the context.
	NewContext(ctx context.Context, fields Fields) context.Context
	// WithContext adds a struct of fields to the log entry.
	WithContext(ctx context.Context) Logger
}

// Fields type to pass when we want to call WithFields for structured logging.
type Fields map[string]interface{}

// Config stores the config for the logger.
type Config struct {
	Type   string `json:"type" envconfig:"LOG_TYPE" required:"true" default:"zap"`
	Level  string `json:"level" envconfig:"LOG_LEVEL" default:"debug"`
	Format string `json:"format" envconfig:"LOG_FORMAT" default:"json"`
}
