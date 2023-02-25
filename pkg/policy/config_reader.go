package policy

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/yaml"
	"github.com/spf13/afero"
)

type ConfigReaderImpl struct {
	fs      afero.Fs
	decoder *yaml.Decoder
}

func NewConfigReader(fs afero.Fs) *ConfigReaderImpl {
	return &ConfigReaderImpl{
		fs:      fs,
		decoder: yaml.NewDecoder(fs),
	}
}

type ConfigReader interface {
	Read([]string) ([]*Config, error)
}

type MockConfigReader struct {
	Cfgs []*Config
	Err  error
}

func (reader *MockConfigReader) Read(files []string) ([]*Config, error) {
	return reader.Cfgs, reader.Err
}

func (reader *ConfigReaderImpl) Read(files []string) ([]*Config, error) {
	policyCfgs := make([]*Config, len(files))
	for i, cfgFilePath := range files {
		policyCfg := &Config{
			Path: cfgFilePath,
			YAML: &ConfigYAML{},
		}
		if err := reader.read(policyCfg); err != nil {
			return nil, fmt.Errorf("read a policy file: %w", err)
		}
		policyCfgs[i] = policyCfg
	}
	return policyCfgs, nil
}

func (reader *ConfigReaderImpl) read(cfg *Config) error {
	if err := reader.decoder.ReadFile(cfg.Path, cfg.YAML); err != nil {
		return fmt.Errorf("parse a policy file as YAML: %w", err)
	}
	if err := cfg.Init(); err != nil {
		return err
	}
	return nil
}

func ParseEnv(env string) []string {
	src := filepath.SplitList(env)
	paths := make([]string, 0, len(src))
	m := make(map[string]struct{}, len(src))
	for _, s := range src {
		if s == "" {
			continue
		}
		if _, ok := m[s]; ok {
			continue
		}
		m[s] = struct{}{}
		paths = append(paths, s)
	}
	return paths
}
