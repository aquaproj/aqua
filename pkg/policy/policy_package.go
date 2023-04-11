package policy

import (
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func getDefaultPolicy() ([]*Config, error) {
	// https://github.com/aquaproj/aqua/issues/1404
	// If no policy file is set, only standard registry is allowed by default.
	cfg := &Config{
		YAML: &ConfigYAML{
			Registries: []*Registry{
				{
					Type: "standard",
				},
			},
			Packages: []*Package{
				{
					RegistryName: "standard",
				},
			},
		},
	}
	if err := cfg.Init(); err != nil {
		return nil, err
	}
	return []*Config{
		cfg,
	}, nil
}

func (pc *Checker) ValidatePackage(logE *logrus.Entry, pkg *config.Package, policies []*Config) error {
	if pc.disabled {
		return nil
	}
	if len(policies) == 0 {
		a, err := getDefaultPolicy()
		if err != nil {
			return err
		}
		policies = a
	}
	for _, policyCfg := range policies {
		policyCfg := policyCfg
		if err := pc.validatePackage(logE, &paramValidatePackage{
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

func (pc *Checker) validatePackage(logE *logrus.Entry, param *paramValidatePackage) error {
	if param.PolicyConfig == nil {
		return nil
	}
	for _, policyPkg := range param.PolicyConfig.Packages {
		f, err := pc.matchPkg(param.Pkg, policyPkg)
		if err != nil {
			logerr.WithError(logE, err).Debug("check if the package matches with a policy")
			continue
		}
		if f {
			return nil
		}
	}
	return errUnAllowedPackage
}

func (pc *Checker) matchPkg(pkg *config.Package, policyPkg *Package) (bool, error) {
	if policyPkg.Name != "" && pkg.Package.Name != policyPkg.Name {
		return false, nil
	}
	if policyPkg.Version != "" {
		sv := pkg.Package.Version
		if pkg.PackageInfo.VersionPrefix != nil {
			sv = strings.TrimPrefix(pkg.Package.Version, *pkg.PackageInfo.VersionPrefix)
		}
		matched, err := expr.EvaluateVersionConstraints(policyPkg.Version, pkg.Package.Version, sv)
		if err != nil {
			return false, fmt.Errorf("evaluate the version constraint of package: %w", err)
		}
		if !matched {
			return false, nil
		}
	}
	return pc.matchRegistry(pkg.Registry, policyPkg.Registry)
}
