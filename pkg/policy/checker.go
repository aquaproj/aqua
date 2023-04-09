package policy

import (
	"errors"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var errUnAllowedPackage = logerr.WithFields(errors.New("this package isn't allowed"), logrus.Fields{
	"doc": "https://aquaproj.github.io/docs/reference/codes/002",
})

type Checker struct {
	disabled bool
}

func NewChecker(param *config.Param) *Checker {
	return &Checker{
		disabled: param.DisablePolicy,
	}
}

// type Checker interface {
// 	ValidatePackage(pkg *config.Package, policies []*Config) error
// }
//
// type MockChecker struct {
// 	Err error
// }
//
// func (pc *MockChecker) ValidatePackage(pkg *config.Package, policies []*Config) error {
// 	return pc.Err
// }
