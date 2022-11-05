package reader

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func NewPolicy(fs afero.Fs) *PolicyConfigReader {
	return &PolicyConfigReader{
		fs: fs,
	}
}

type PolicyConfigReader struct {
	fs afero.Fs
}

func (reader *PolicyConfigReader) Read(configFilePath string, cfg *policy.Config) error {
	file, err := reader.fs.Open(configFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
	}
	return nil
}
