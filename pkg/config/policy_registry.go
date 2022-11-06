package config

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/expr"
)

func (pc *PolicyChecker) ValidateRegistry(rgst *aqua.Registry, policyConfig *policy.Config) error {
	if policyConfig == nil {
		return errUnAllowedPackage
	}
	for _, regist := range policyConfig.Registries {
		f, err := pc.matchRegistry(rgst, regist)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedRegistry
}

func (pc *PolicyChecker) matchRegistry(rgst *aqua.Registry, rgstPolicy *policy.Registry) (bool, error) {
	if rgst.Type != rgstPolicy.Type {
		return false, nil
	}
	if rgstPolicy.Ref != "" {
		matched, err := expr.EvaluateVersionConstraints(rgstPolicy.Ref, rgst.Ref)
		if err != nil {
			return false, fmt.Errorf("evaluate the version constraint of registry: %w", err)
		}
		return matched, nil
	}
	return true, nil
}
