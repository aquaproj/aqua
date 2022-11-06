package domain

import (
	"github.com/aquaproj/aqua/pkg/config"
)

type PolicyChecker interface {
	ValidatePackage(param *config.ParamValidatePackage) error
	ValidateRegistry(param *config.ParamValidateRegistry) error
}

type MockPolicyChecker struct {
	Err error
}

func (pc *MockPolicyChecker) ValidatePackage(param *config.ParamValidatePackage) error {
	return pc.Err
}

func (pc *MockPolicyChecker) ValidateRegistry(param *config.ParamValidateRegistry) error {
	return pc.Err
}
