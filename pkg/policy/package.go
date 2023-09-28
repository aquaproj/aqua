package policy

import (
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func ValidatePackage(logE *logrus.Entry, pkg *config.Package, policies []*Config) error {
	if len(policies) == 0 {
		a, err := getDefaultPolicy()
		if err != nil {
			return err
		}
		policies = a
	}
	for _, policyCfg := range policies {
		if err := validatePackage(logE, &paramValidatePackage{
			Pkg:          pkg,
			PolicyConfig: policyCfg.YAML,
		}); err == nil {
			return nil
		}
	}
	return errUnAllowedPackage
}

type paramValidatePackage struct {
	Pkg          *config.Package
	PolicyConfig *ConfigYAML
}

func validatePackage(logE *logrus.Entry, param *paramValidatePackage) error {
	if param.PolicyConfig == nil {
		return nil
	}
	for _, policyPkg := range param.PolicyConfig.Packages {
		f, err := matchPkg(param.Pkg, policyPkg)
		if err != nil {
			// If it fails to check if the policy matches with the package, output a debug log and treat as the policy doesn't match with the package.
			logerr.WithError(logE, err).Debug("check if the package matches with a policy")
			continue
		}
		if f {
			return nil
		}
	}
	return errUnAllowedPackage
}

func matchPkg(pkg *config.Package, policyPkg *Package) (bool, error) {
	if policyPkg.Name != "" && pkg.Package.Name != policyPkg.Name {
		return false, nil
	}
	if policyPkg.Version != "" {
		sv := pkg.Package.Version
		if pkg.PackageInfo.VersionPrefix != "" {
			sv = strings.TrimPrefix(pkg.Package.Version, pkg.PackageInfo.VersionPrefix)
		}
		matched, err := expr.EvaluateVersionConstraints(policyPkg.Version, pkg.Package.Version, sv)
		if err != nil {
			return false, fmt.Errorf("evaluate the version constraint of package: %w", err)
		}
		if !matched {
			return false, nil
		}
	}
	return matchRegistry(pkg.Registry, policyPkg.Registry)
}

func matchRegistry(rgst *aqua.Registry, rgstPolicy *Registry) (bool, error) {
	if rgst.Type != rgstPolicy.Type {
		return false, nil
	}
	if rgst.Type == "local" {
		return rgst.Path == rgstPolicy.Path, nil
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
		matched, err := expr.EvaluateVersionConstraints(rgstPolicy.Ref, rgst.Ref, rgst.Ref)
		if err != nil {
			return false, fmt.Errorf("evaluate the version constraint of registry: %w", err)
		}
		return matched, nil
	}
	return true, nil
}
