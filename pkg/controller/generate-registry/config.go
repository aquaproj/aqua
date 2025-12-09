package genrgst

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	VersionPrefix   string
	VersionFilter   *vm.Program
	AllAssetsFilter *vm.Program
	Package         string
}

type RawConfig struct {
	VersionFilter   string `yaml:"version_filter" json:"version_filter,omitempty"`
	VersionPrefix   string `yaml:"version_prefix" json:"version_prefix,omitempty"`
	AllAssetsFilter string `yaml:"all_assets_filter" json:"all_assets_filter,omitempty"`
	Package         string `yaml:"name" json:"name"`
}

func (c *Config) FromRaw(raw *RawConfig) error {
	if raw == nil {
		return nil
	}

	c.Package = raw.Package
	c.VersionPrefix = raw.VersionPrefix

	if raw.VersionFilter != "" {
		r, err := expr.CompileVersionFilter(raw.VersionFilter)
		if err != nil {
			return fmt.Errorf("compile a version expression: %w", err)
		}
		c.VersionFilter = r
	}

	if raw.AllAssetsFilter != "" {
		a, err := expr.CompileAssetFilter(raw.AllAssetsFilter)
		if err != nil {
			return fmt.Errorf("compile an asset expression: %w", err)
		}
		c.AllAssetsFilter = a
	}

	return nil
}

func readConfig(fs afero.Fs, path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	f, err := fs.Open(path)
	if err != nil {
		return fmt.Errorf("open a generate configuration file: %w", err)
	}
	defer f.Close()
	raw := &RawConfig{}
	if err := yaml.NewDecoder(f).Decode(raw); err != nil {
		return fmt.Errorf("decode a generate configuration file as YAML: %w", err)
	}
	return cfg.FromRaw(raw)
}
