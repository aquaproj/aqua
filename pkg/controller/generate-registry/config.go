package genrgst

import (
	"fmt"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	VersionPrefix   string
	VersionFilter   *vm.Program
	AllAssetsFilter *vm.Program
	Package         string
	// Spellings names the platforms a release writes in a way the parser doesn't
	// know. Without them an asset belongs to no platform, and the package is
	// generated without it rather than with a wrong name for it.
	Spellings asset.Spellings
}

type RawConfig struct {
	VersionFilter   string `yaml:"version_filter" json:"version_filter,omitempty"`
	VersionPrefix   string `yaml:"version_prefix" json:"version_prefix,omitempty"`
	AllAssetsFilter string `yaml:"all_assets_filter" json:"all_assets_filter,omitempty"`
	Package         string `yaml:"name" json:"name"`
	// Replacements says how this release writes a platform, for the ones the
	// parser doesn't recognise: luau-lang/luau calls its Linux build
	// luau-ubuntu.zip. It is the same map a registry writes, so a caller that
	// already has the package's definition can hand it over as it is.
	Replacements map[string]string `yaml:"replacements" json:"replacements,omitempty"`
}

func (c *Config) FromRaw(raw *RawConfig) error {
	if raw == nil {
		return nil
	}

	c.Package = raw.Package
	c.VersionPrefix = raw.VersionPrefix
	c.Spellings = raw.Replacements

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

func readConfig(path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
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
