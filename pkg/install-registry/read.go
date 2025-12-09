package registry

import (
	"encoding/json"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"go.yaml.in/yaml/v2"
)

func (is *Installer) readYAMLRegistry(p string, registry *registry.Config) error {
	f, err := is.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(registry); err != nil {
		return fmt.Errorf("parse the registry configuration as YAML: %w", err)
	}
	return nil
}

func (is *Installer) readJSONRegistry(p string, registry *registry.Config) error {
	f, err := is.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(registry); err != nil {
		return fmt.Errorf("parse the registry configuration as JSON: %w", err)
	}
	return nil
}

func (is *Installer) readRegistry(p string, registry *registry.Config) error {
	if isJSON(p) {
		return is.readJSONRegistry(p, registry)
	}
	return is.readYAMLRegistry(p, registry)
}
