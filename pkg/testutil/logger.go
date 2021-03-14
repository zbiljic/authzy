package testutil

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/zbiljic/authzy/pkg/logger"
)

// Compile-time proof of interface implementation.
var _ logger.Logger = (*Logger)(nil)

// Logger defines a logging structure for plugins.
type Logger struct {
	Name string // Name is the plugin name, will be printed in the `[]`.

	fields logger.Fields
}

func (l Logger) sprint(args ...interface{}) (out []interface{}) {
	out = append(out, fmt.Sprint(args...))
	if len(l.fields) > 0 {
		for k, v := range l.fields {
			out = append(out, fmt.Sprintf(" %s=%v", k, v))
		}
	}

	return
}

func (l Logger) sprintf(format string, args ...interface{}) (out []interface{}) {
	out = append(out, fmt.Sprintf(format, args...))
	if len(l.fields) > 0 {
		for k, v := range l.fields {
			out = append(out, fmt.Sprintf(" %s=%v", k, v))
		}
	}

	return
}

// SetOutput sets the output destination for the logger.
func (l Logger) SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// Debug logs a debug message, patterned after log.Print.
func (l Logger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! [" + l.Name + "] "}, l.sprint(args...)...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l Logger) Debugf(format string, args ...interface{}) {
	log.Print(l.sprintf("D! ["+l.Name+"] "+format, args...)...)
}

// Info logs an information message, patterned after log.Print.
func (l Logger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! [" + l.Name + "] "}, l.sprint(args...)...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l Logger) Infof(format string, args ...interface{}) {
	log.Print(l.sprintf("I! ["+l.Name+"] "+format, args...)...)
}

// Warn logs a warning message, patterned after log.Print.
func (l Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, l.sprint(args...)...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l Logger) Warnf(format string, args ...interface{}) {
	log.Print(l.sprintf("W! ["+l.Name+"] "+format, args...)...)
}

// Error logs an error message, patterned after log.Print.
func (l Logger) Error(args ...interface{}) {
	log.Print(append([]interface{}{"E! [" + l.Name + "] "}, l.sprint(args...)...)...)
}

// Errorf logs an error message, patterned after log.Printf.
func (l Logger) Errorf(format string, args ...interface{}) {
	log.Print(l.sprintf("E! ["+l.Name+"] "+format, args...)...)
}

func (l Logger) WithFields(fields logger.Fields) logger.Logger {
	return &Logger{Name: l.Name, fields: fields}
}

// fieldsContextKey is used to reference logger fields as context value.
type fieldsContextKey struct{}

func (l Logger) NewContext(ctx context.Context, fields logger.Fields) context.Context {
	if ctxFields, ok := ctx.Value(fieldsContextKey{}).(logger.Fields); ok {
		// extend context fields
		for k, v := range fields {
			ctxFields[k] = v
		}
		return context.WithValue(ctx, fieldsContextKey{}, ctxFields)
	}

	return context.WithValue(ctx, fieldsContextKey{}, fields)
}

func (l Logger) WithContext(ctx context.Context) logger.Logger {
	if ctxFields, ok := ctx.Value(fieldsContextKey{}).(logger.Fields); ok {
		return l.WithFields(ctxFields)
	}
	return l
}
