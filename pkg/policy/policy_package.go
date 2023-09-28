package policy

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
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

type paramValidatePackage struct {
	Pkg          *config.Package
	PolicyConfig *ConfigYAML
}
