package reader

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config/security"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func NewSecurity(fs afero.Fs) *SecurityConfigReader {
	return &SecurityConfigReader{
		fs: fs,
	}
}

type SecurityConfigReader struct {
	fs afero.Fs
}

func (reader *SecurityConfigReader) Read(configFilePath string, cfg *security.Config) error {
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
