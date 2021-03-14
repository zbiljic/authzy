package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/logger/zlogger"
)

type loggerFxPrinter struct {
	log logger.Logger
}

func NewFxLogger(log ...logger.Logger) fx.Printer {
	if len(log) == 0 {
		log = append(log, zlogger.Instance())
	}
	return loggerFxPrinter{log[0]}
}

func (l loggerFxPrinter) Printf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (loggerFxPrinter) String() string {
	return "zlogger"
}
