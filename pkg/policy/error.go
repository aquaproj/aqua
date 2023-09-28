package policy

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	ErrConfigFileNotFound = errors.New("policy file isn't found")
	errUnAllowedPackage   = logerr.WithFields(errors.New("this package isn't allowed"), logrus.Fields{
		"doc": "https://aquaproj.github.io/docs/reference/codes/002",
	})
)
