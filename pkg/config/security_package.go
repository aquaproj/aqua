package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/pkg/config/security"
	"github.com/aquaproj/aqua/pkg/expr"
)

func (sec *SecurityChecker) validatePkgs(pkg *Package, pkgs []*security.Package) error {
	for _, secPkg := range pkgs {
		f, err := sec.matchPkg(pkg, secPkg)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedPackage
}

func (sec *SecurityChecker) matchPkg(pkg *Package, secPkg *security.Package) (bool, error) {
	f, err := sec.matchPkgID(pkg, secPkg)
	if err != nil {
		return false, err
	}
	if !f {
		return false, nil
	}
	return sec.matchPkgVersion(pkg, secPkg)
}

func (sec *SecurityChecker) matchPkgID(pkg *Package, secPkg *security.Package) (bool, error) {
	pkgID, err := pkg.GetChecksumID(sec.rt)
	if err != nil {
		return false, err
	}
	switch secPkg.IDFormat {
	case "regexp":
		matched, err := regexp.MatchString(secPkg.ID, pkgID)
		if err != nil {
			return false, fmt.Errorf("match the package id with regular expression: %w", err)
		}
		if matched {
			return true, nil
		}
		return false, nil
	case "glob":
		matched, err := filepath.Match(secPkg.ID, pkgID)
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

func (sec *SecurityChecker) matchPkgVersion(pkg *Package, secPkg *security.Package) (bool, error) {
	matched, err := expr.EvaluateVersionConstraints(secPkg.Version, pkg.Package.Version)
	if err != nil {
		return false, fmt.Errorf("evaluate the version constraint of package: %w", err)
	}
	return matched, nil
}
