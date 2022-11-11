package policy

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func NewConfigReader(fs afero.Fs) *ConfigReader {
	return &ConfigReader{
		fs: fs,
	}
}

type ConfigReader struct {
	fs afero.Fs
}

func (reader *ConfigReader) Read(cfg *Config) error {
	file, err := reader.fs.Open(cfg.Path)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg.YAML); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", cfg.Path, err)
	}
	if err := cfg.Init(); err != nil {
		return err
	}
	return nil
}
