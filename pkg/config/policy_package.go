package config

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/expr"
)

type ParamValidatePackage struct {
	Pkg           *Package
	PolicyConfig  *policy.Config
	ConfigFileDir string
	PolicyFileDir string
}

func (pc *PolicyChecker) ValidatePackage(param *ParamValidatePackage) error {
	if param.PolicyConfig == nil {
		return nil
	}
	for _, policyPkg := range param.PolicyConfig.Packages {
		f, err := pc.matchPkg(param.Pkg, policyPkg, param.ConfigFileDir, param.PolicyFileDir)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedPackage
}

func (pc *PolicyChecker) matchPkg(pkg *Package, policyPkg *policy.Package, cfgDir, policyDir string) (bool, error) {
	if pkg.Package.Name != policyPkg.Name {
		return false, nil
	}
	if policyPkg.Version != "" {
		matched, err := expr.EvaluateVersionConstraints(policyPkg.Version, pkg.Package.Version)
		if err != nil {
			return false, fmt.Errorf("evaluate the version constraint of package: %w", err)
		}
		if !matched {
			return false, nil
		}
	}
	return pc.matchRegistry(pkg.Registry, policyPkg.Registry, cfgDir, policyDir)
}
