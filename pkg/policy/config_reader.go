package policy

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type ConfigReaderImpl struct {
	fs afero.Fs
}

func NewConfigReader(fs afero.Fs) *ConfigReaderImpl {
	return &ConfigReaderImpl{
		fs: fs,
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
	if len(files) == 0 {
		return reader.readDefault()
	}
	policyCfgs := make([]*Config, len(files))
	for i, cfgFilePath := range files {
		policyCfg := &Config{
			Path: cfgFilePath,
			YAML: &ConfigYAML{},
		}
		if err := reader.read(policyCfg); err != nil {
			return nil, fmt.Errorf("read the policy config file: %w", err)
		}
		policyCfgs[i] = policyCfg
	}
	return policyCfgs, nil
}

func (reader *ConfigReaderImpl) readDefault() ([]*Config, error) {
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
	if err := cfg.Init(); err != nil {
		return nil, err
	}
	return []*Config{
		cfg,
	}, nil
}

func (reader *ConfigReaderImpl) read(cfg *Config) error {
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
