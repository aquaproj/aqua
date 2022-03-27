package log

import (
	"github.com/sirupsen/logrus"
)

func New() *logrus.Entry {
	return logrus.WithField("program", "aqua")
}

type Logger struct {
	version string
}

func NewLogger(v string) *Logger {
	return &Logger{
		version: v,
	}
}

func (logger *Logger) LogE() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"aqua_version": logger.version,
		"program":      "aqua",
	})
}
