package mailer

import (
	"io"

	"github.com/sirupsen/logrus"

	"github.com/zbiljic/authzy/pkg/logger"
)

type logrusLogger struct {
	*logrus.Logger
}

func newLogrusLogger(logger logger.Logger) logrus.FieldLogger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	l.AddHook(&loggerHook{logger})
	return &logrusLogger{Logger: l}
}

type loggerHook struct {
	log logger.Logger
}

func (h *loggerHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *loggerHook) Fire(e *logrus.Entry) error {
	switch e.Level {
	case logrus.ErrorLevel:
		h.log.WithFields(logger.Fields(e.Data)).Error(e.Message)
	case logrus.WarnLevel:
		h.log.WithFields(logger.Fields(e.Data)).Warn(e.Message)
	case logrus.InfoLevel:
		h.log.WithFields(logger.Fields(e.Data)).Info(e.Message)
	case logrus.DebugLevel:
		h.log.WithFields(logger.Fields(e.Data)).Debug(e.Message)
	}
	return nil
}
