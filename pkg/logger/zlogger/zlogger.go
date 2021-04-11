package zlogger

import (
	"fmt"

	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/logger/zap"
	"github.com/zbiljic/authzy/pkg/logger/zerolog"
)

func New(config *logger.Config) (logger.Logger, error) {
	var (
		logger logger.Logger
		err    error
	)

	switch config.Type {
	case "zap":
		logger, err = zap.NewLogger(config.Level, config.Format)
		if err != nil {
			return nil, err
		}
	case "zerolog":
		logger, err = zerolog.NewLogger(config.Level, config.Format)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid logger type: %s", config.Type)
	}

	return logger, nil
}
