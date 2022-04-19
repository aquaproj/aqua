package log

import (
	"github.com/sirupsen/logrus"
)

func New(version string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"aqua_version": version,
		"program":      "aqua",
	})
}

func SetLevel(level string, logE *logrus.Entry) {
	if level == "" {
		return
	}
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logE.WithField("log_level", level).WithError(err).Error("the log level is invalid")
		return
	}
	logrus.SetLevel(lvl)
}
