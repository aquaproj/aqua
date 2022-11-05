package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/expr"
)

func (pc *PolicyChecker) validateRegistries(pkg *Package, registries []*policy.Registry) error {
	for _, regist := range registries {
		f, err := pc.matchRegistry(pkg, regist)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedRegistry
}

func (pc *PolicyChecker) matchRegistry(pkg *Package, regist *policy.Registry) (bool, error) {
	f, err := pc.matchRegistryID(pkg, regist)
	if err != nil {
		return false, err
	}
	if !f {
		return false, nil
	}
	return pc.matchRegistryVersion(pkg, regist)
}

func (pc *PolicyChecker) matchRegistryID(pkg *Package, regist *policy.Registry) (bool, error) { //nolint:cyclop
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

func (pc *PolicyChecker) matchRegistryVersion(pkg *Package, regist *policy.Registry) (bool, error) {
	matched, err := expr.EvaluateVersionConstraints(regist.Version, pkg.Registry.Ref)
	if err != nil {
		return false, fmt.Errorf("evaluate the version constraint of registry: %w", err)
	}
	return matched, nil
}
