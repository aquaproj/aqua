package domain

import (
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/policy"
)

type PolicyChecker interface {
	ValidatePackage(pkg *config.Package, cfg *policy.Config) error
	ValidateRegistry(rgst *aqua.Registry, policyConfig *policy.Config) error
}

type MockPolicyChecker struct {
	Err error
}

func (pc *MockPolicyChecker) ValidatePackage(pkg *config.Package, cfg *policy.Config) error {
	return pc.Err
}

func (pc *MockPolicyChecker) ValidateRegistry(rgst *aqua.Registry, policyConfig *policy.Config) error {
	return pc.Err
}
