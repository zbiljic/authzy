package api

import (
	"github.com/gorilla/handlers"

	"github.com/zbiljic/authzy/pkg/logger"
)

// Compile-time proof of interface implementation.
var _ handlers.RecoveryHandlerLogger = (*recoveryHandlerLogger)(nil)

type recoveryHandlerLogger struct {
	log logger.Logger
}

func (l recoveryHandlerLogger) Println(args ...interface{}) {
	l.log.Error(args)
}
