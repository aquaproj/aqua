package policy

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	errUnAllowedPackage = logerr.WithFields(errors.New("this package isn't allowed"), logrus.Fields{
		"doc": "https://github.com/aquaproj/aqua/issues/1306", // TODO change URL
	})
	errUnAllowedRegistry = errors.New("this registry isn't allowed")
)

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}
