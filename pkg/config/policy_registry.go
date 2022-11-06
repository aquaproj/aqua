package config

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/expr"
)

type ParamValidateRegistry struct {
	Registry      *aqua.Registry
	PolicyConfig  *policy.Config
	ConfigFileDir string
	PolicyFileDir string
}

func (pc *PolicyChecker) ValidateRegistry(param *ParamValidateRegistry) error {
	if param.PolicyConfig == nil {
		return errUnAllowedPackage
	}
	for _, regist := range param.PolicyConfig.Registries {
		f, err := pc.matchRegistry(param.Registry, regist, param.ConfigFileDir, param.PolicyFileDir)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return errUnAllowedRegistry
}

func (pc *PolicyChecker) matchRegistry(rgst *aqua.Registry, rgstPolicy *policy.Registry, cfgDir, policyDir string) (bool, error) {
	if rgst.Type != rgstPolicy.Type {
		return false, nil
	}
	if rgst.Type == "local" {
		return pc.matchLocalRegistryPath(cfgDir, rgst.Path, policyDir, rgstPolicy.Path)
	}
	if rgst.RepoOwner != rgstPolicy.RepoOwner {
		return false, nil
	}
	if rgst.RepoName != rgstPolicy.RepoName {
		return false, nil
	}
	if rgst.Path != rgstPolicy.Path {
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

func (pc *PolicyChecker) matchLocalRegistryPath(cfgDir, rgstPath, policyDir, rgstPolicyPath string) (bool, error) {
	if !filepath.IsAbs(rgstPath) {
		rgstPath = filepath.Join(cfgDir, rgstPath)
	}
	if !filepath.IsAbs(rgstPolicyPath) {
		rgstPolicyPath = filepath.Join(policyDir, rgstPolicyPath)
	}
	return rgstPath == rgstPolicyPath, nil
}
