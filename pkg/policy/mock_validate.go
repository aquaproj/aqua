package policy

import (
	"github.com/sirupsen/logrus"
)

type MockValidator struct {
	Err error
}

func (v *MockValidator) Allow(p string) error {
	return v.Err
}

func (v *MockValidator) Deny(p string) error {
	return v.Err
}

func (v *MockValidator) Validate(p string) error {
	return v.Err
}

func (v *MockValidator) Warn(logE *logrus.Entry, policyFilePath string, updated bool) error {
	return v.Err
}
