package genrgst

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type Config struct {
	VersionPrefix   string
	VersionFilter   *vm.Program
	AllAssetsFilter *vm.Program
	Package         string
}

type RawConfig struct {
	VersionFilter   string `json:"version_filter,omitempty"    yaml:"version_filter"`
	VersionPrefix   string `json:"version_prefix,omitempty"    yaml:"version_prefix"`
	AllAssetsFilter string `json:"all_assets_filter,omitempty" yaml:"all_assets_filter"`
	Package         string `json:"name"                        yaml:"name"`
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
