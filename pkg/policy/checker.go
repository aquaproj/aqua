package policy

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var errUnAllowedPackage = logerr.WithFields(errors.New("this package isn't allowed"), logrus.Fields{
	"doc": "https://aquaproj.github.io/docs/reference/codes/002",
})

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}
