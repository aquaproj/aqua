package config

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/expr"
)

func (pc *PolicyChecker) ValidatePackage(pkg *Package, cfg *policy.Config) error {
	if cfg == nil {
		return nil
	}
	for _, policyPkg := range cfg.Packages {
		f, err := pc.matchPkg(pkg, policyPkg)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedPackage
}

func (pc *PolicyChecker) matchPkg(pkg *Package, policyPkg *policy.Package) (bool, error) {
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
	return pc.matchRegistry(pkg.Registry, policyPkg.Registry)
}
