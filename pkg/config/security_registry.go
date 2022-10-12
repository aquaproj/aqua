package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/pkg/config/security"
	"github.com/aquaproj/aqua/pkg/expr"
)

func (sec *SecurityChecker) validateRegistries(pkg *Package, registries []*security.Registry) error {
	for _, regist := range registries {
		f, err := sec.matchRegistry(pkg, regist)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedRegistry
}

func (sec *SecurityChecker) matchRegistry(pkg *Package, regist *security.Registry) (bool, error) {
	f, err := sec.matchRegistryID(pkg, regist)
	if err != nil {
		return false, err
	}
	if !f {
		return false, nil
	}
	return sec.matchRegistryVersion(pkg, regist)
}

func (sec *SecurityChecker) matchRegistryID(pkg *Package, regist *security.Registry) (bool, error) { //nolint:cyclop
	registID := pkg.Registry.GetID()
	switch regist.ID {
	case "standard":
		if pkg.Registry.Type == "github_content" && pkg.Registry.RepoOwner == "aquaproj" && pkg.Registry.RepoName == "aqua" {
			return true, nil
		}
		return false, nil
	case "local":
		if pkg.Registry.Type == "local" {
			return true, nil
		}
		return false, nil
	}
	switch regist.IDFormat {
	case "regexp":
		matched, err := regexp.MatchString(regist.ID, registID)
		if err != nil {
			return false, fmt.Errorf("match the registry id with regular expression: %w", err)
		}
		if matched {
			return true, nil
		}
		return false, nil
	case "glob":
		matched, err := filepath.Match(regist.ID, registID)
		if err != nil {
			return false, fmt.Errorf("match the registry id with glob: %w", err)
		}
		if matched {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}

func (sec *SecurityChecker) matchRegistryVersion(pkg *Package, regist *security.Registry) (bool, error) {
	matched, err := expr.EvaluateVersionConstraints(regist.Version, pkg.Registry.Ref)
	if err != nil {
		return false, fmt.Errorf("evaluate the version constraint of registry: %w", err)
	}
	return matched, nil
}
