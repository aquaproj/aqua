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
	Read(policyConfigFiles []string) ([]*Config, error)
	ReadFile(policyConfigFile string) (*Config, error)
}

type MockConfigReader struct {
	Cfgs []*Config
	Err  error
}

func (r *MockConfigReader) Read(files []string) ([]*Config, error) {
	return r.Cfgs, r.Err
}

func (r *ConfigReaderImpl) Read(files []string) ([]*Config, error) {
	policyCfgs := make([]*Config, len(files))
	for i, cfgFilePath := range files {
		policyCfg := &Config{
			Path: cfgFilePath,
			YAML: &ConfigYAML{},
		}
		if err := r.read(policyCfg); err != nil {
			return nil, fmt.Errorf("read the policy config file: %w", err)
		}
		policyCfgs[i] = policyCfg
	}
	return policyCfgs, nil
}

func (r *ConfigReaderImpl) ReadFile(file string) (*Config, error) {
	policyCfg := &Config{
		Path: file,
		YAML: &ConfigYAML{},
	}
	if err := r.read(policyCfg); err != nil {
		return nil, fmt.Errorf("read the policy config file: %w", err)
	}
	return policyCfg, nil
}

func (r *ConfigReaderImpl) read(cfg *Config) error {
	file, err := r.fs.Open(cfg.Path)
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
