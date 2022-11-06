package domain

import (
	"github.com/aquaproj/aqua/pkg/policy"
)

type PolicyChecker interface {
	ValidatePackage(param *policy.ParamValidatePackage) error
	ValidateRegistry(param *policy.ParamValidateRegistry) error
}

type MockPolicyChecker struct {
	Err error
}

func (pc *MockPolicyChecker) ValidatePackage(param *policy.ParamValidatePackage) error {
	return pc.Err
}

func (pc *MockPolicyChecker) ValidateRegistry(param *policy.ParamValidateRegistry) error {
	return pc.Err
}
