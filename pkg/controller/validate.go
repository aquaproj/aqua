package controller

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var validate = validator.New() //nolint:gochecknoglobals

func validateRegistries(registries Registries) error {
	for _, registry := range registries {
		if err := registry.validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateRegistryContent(registryContent *RegistryContent) error {
	return validatePackageInfos(registryContent.PackageInfos)
}

func validateConfig(cfg *Config) error {
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}
	if err := validatePackages(cfg.Packages); err != nil {
		return err
	}
	if cfg.InlineRegistry != nil {
		if err := validateRegistryContent(cfg.InlineRegistry); err != nil {
			return err
		}
	}
	if err := validateRegistries(cfg.Registries); err != nil {
		return err
	}
	return nil
}

func validatePackages(pkgs []*Package) error {
	return nil
}

func validatePackageInfos(pkgInfos PackageInfos) error {
	names := make(map[string]struct{}, len(pkgInfos))
	for _, pkgInfo := range pkgInfos {
		name := pkgInfo.GetName()
		if _, ok := names[name]; ok {
			return logerr.WithFields(errPkgNameMustBeUniqueInRegistry, logrus.Fields{ //nolint:wrapcheck
				"package_name": name,
			})
		}
		names[name] = struct{}{}
	}
	return nil
}
