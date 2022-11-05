package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/expr"
)

func (pc *PolicyChecker) validatePkgs(pkg *Package, pkgs []*policy.Package) error {
	for _, policyPkg := range pkgs {
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
	f, err := pc.matchPkgID(pkg, policyPkg)
	if err != nil {
		return false, err
	}
	if !f {
		return false, nil
	}
	return pc.matchPkgVersion(pkg, policyPkg)
}

func (pc *PolicyChecker) matchPkgID(pkg *Package, policyPkg *policy.Package) (bool, error) {
	pkgID, err := pkg.GetChecksumID(pc.rt)
	if err != nil {
		return false, err
	}
	switch policyPkg.IDFormat {
	case "regexp":
		matched, err := regexp.MatchString(policyPkg.ID, pkgID)
		if err != nil {
			return false, fmt.Errorf("match the package id with regular expression: %w", err)
		}
		if matched {
			return true, nil
		}
		return false, nil
	case "glob":
		matched, err := filepath.Match(policyPkg.ID, pkgID)
		if err != nil {
			return false, fmt.Errorf("match the package id with glob: %w", err)
		}
		if matched {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}

func (pc *PolicyChecker) matchPkgVersion(pkg *Package, policyPkg *policy.Package) (bool, error) {
	matched, err := expr.EvaluateVersionConstraints(policyPkg.Version, pkg.Package.Version)
	if err != nil {
		return false, fmt.Errorf("evaluate the version constraint of package: %w", err)
	}
	return matched, nil
}
