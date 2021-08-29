package log

import (
	"github.com/sirupsen/logrus"
)

func New() *logrus.Entry {
	return logrus.WithField("program", "aqua")
}
