//go:build windows
// +build windows

package log

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(colorable.NewColorableStderr())
}
