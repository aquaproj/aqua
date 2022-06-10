package validate

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New() //nolint:gochecknoglobals

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
