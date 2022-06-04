package validate

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var validate = validator.New() //nolint:gochecknoglobals

func RegistryConfig(registryContent *registry.Config) error {
	return validatePackageInfos(registryContent.PackageInfos)
}

func Config(cfg *aqua.Config) error {
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}
	if err := validatePackages(cfg.Packages); err != nil {
		return err
	}
	if err := validateRegistries(cfg.Registries); err != nil {
		return err
	}
	return nil
}

func validateRegistries(registries aqua.Registries) error {
	for _, registry := range registries {
		if err := registry.Validate(); err != nil {
			return err //nolint:wrapcheck
		}
	}
	return nil
}

func validatePackages(pkgs []*aqua.Package) error {
	return nil
}

func validatePackageInfos(pkgInfos registry.PackageInfos) error {
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
