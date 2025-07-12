package policy

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
	err := cfg.Init()
	if err != nil {
		return nil, err
	}

	return []*Config{
		cfg,
	}, nil
}
