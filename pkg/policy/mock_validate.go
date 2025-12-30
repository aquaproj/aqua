package policy

import (
	"log/slog"
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

func (v *MockValidator) Warn(logger *slog.Logger, policyFilePath string, updated bool) error {
	return v.Err
}
