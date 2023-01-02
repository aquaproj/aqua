package policy

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var errUnAllowedPackage = logerr.WithFields(errors.New("this package isn't allowed"), logrus.Fields{
	"doc": "https://aquaproj.github.io/docs/reference/codes/002",
})

type CheckerImpl struct{}

func NewChecker() *CheckerImpl {
	return &CheckerImpl{}
}

type Checker interface {
	ValidatePackage(param *ParamValidatePackage) error
}

type MockChecker struct {
	Err error
}

func (pc *MockChecker) ValidatePackage(param *ParamValidatePackage) error {
	return pc.Err
}
