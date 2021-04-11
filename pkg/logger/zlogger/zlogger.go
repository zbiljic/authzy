package zlogger

import (
	"fmt"

	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/logger/zap"
	"github.com/zbiljic/authzy/pkg/logger/zerolog"
)

func New(config *logger.Config) (logger.Logger, error) {
	switch config.Type {
	case "zap":
		return zap.NewLogger(config.Level, config.Format)
	case "zerolog":
		return zerolog.NewLogger(config.Level, config.Format)
	default:
		return nil, fmt.Errorf("invalid logger type: %s", config.Type)
	}
}
