package genrgst

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version *vm.Program
	Asset   *vm.Program
}

type RawConfig struct {
	Version string `json:"version,omitempty"`
	Asset   string `json:"asset,omitempty"`
}

func (c *Config) FromRaw(raw *RawConfig) error {
	if raw == nil {
		return nil
	}

	r, err := expr.CompileVersionFilter(raw.Version)
	if err != nil {
		return fmt.Errorf("compile a version expression: %w", err)
	}
	c.Version = r

	a, err := expr.CompileAssetFilter(raw.Asset)
	if err != nil {
		return fmt.Errorf("compile an asset expression: %w", err)
	}
	c.Asset = a

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
