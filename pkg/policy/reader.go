package policy

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Reader interface {
	Read(logE *logrus.Entry, policyFilePath, aquaConfigDir string) (*Config, error)
	ReadFromEnv(policyFilePaths []string) ([]*Config, error)
	ValidateAndRead(logE *logrus.Entry, policyFilePath string) (*Config, error)
}

type MockReader struct {
	Config  *Config
	Configs []*Config
	Err     error
}

func (reader *MockReader) ReadFromEnv(policyFilePaths []string) ([]*Config, error) {
	allowCfgs(reader.Configs)
	return reader.Configs, reader.Err
}

func (reader *MockReader) Read(logE *logrus.Entry, policyFilePath, aquaConfigDir string) (*Config, error) {
	return reader.Config, reader.Err
}

func (reader *MockReader) ValidateAndRead(logE *logrus.Entry, policyFilePath string) (*Config, error) {
	return reader.Config, reader.Err
}

type ReaderImpl struct {
	mutex     *sync.RWMutex
	policies  map[string]*Config
	fs        afero.Fs
	validator Validator
	finder    ConfigFinder
	reader    ConfigReader
}

func NewReader(fs afero.Fs, validator Validator, finder ConfigFinder, reader ConfigReader) *ReaderImpl {
	return &ReaderImpl{
		mutex:     &sync.RWMutex{},
		policies:  map[string]*Config{},
		fs:        fs,
		validator: validator,
		finder:    finder,
		reader:    reader,
	}
}

func (reader *ReaderImpl) get(p string) *Config {
	reader.mutex.RLock()
	defer reader.mutex.RUnlock()
	return reader.policies[p]
}

func (reader *ReaderImpl) set(p string, cfg *Config) {
	reader.mutex.Lock()
	defer reader.mutex.Unlock()
	reader.policies[p] = cfg
}

func allowCfgs(cfgs []*Config) {
	for _, cfg := range cfgs {
		cfg.Allowed = true
	}
}

func (reader *ReaderImpl) ReadFromEnv(policyFilePaths []string) ([]*Config, error) {
	cfgs, err := reader.reader.Read(policyFilePaths)
	if err != nil {
		return nil, fmt.Errorf("read policies from the environment variable: %w", err)
	}
	allowCfgs(cfgs)
	return cfgs, nil
}

func (reader *ReaderImpl) Read(logE *logrus.Entry, policyFilePath, aquaConfigDir string) (*Config, error) {
	if policyFilePath != "" {
		f, err := afero.Exists(reader.fs, policyFilePath)
		if err != nil {
			return nil, fmt.Errorf("check if a policy file exists: %w", err)
		}
		if !f {
			return nil, ErrConfigFileNotFound
		}
		return reader.ValidateAndRead(logE, policyFilePath)
	}
	policyFilePath, err := reader.finder.Find("", aquaConfigDir)
	if err != nil {
		return nil, fmt.Errorf("find a policy file: %w", err)
	}
	return reader.ValidateAndRead(logE, policyFilePath)
}

func (reader *ReaderImpl) ValidateAndRead(logE *logrus.Entry, policyFilePath string) (*Config, error) {
	if cfg := reader.get(policyFilePath); cfg != nil {
		if cfg.Allowed {
			return cfg, nil
		}
		return nil, nil //nolint:nilnil
	}
	if err := reader.validator.Validate(policyFilePath); err != nil {
		reader.set(policyFilePath, &Config{})
		if err := reader.validator.Warn(logE, policyFilePath); err != nil {
			logE.WithError(err).Warn("warn an denied policy file")
		}
		return nil, nil //nolint:nilnil
	}
	cfg, err := reader.reader.ReadFile(policyFilePath)
	if err != nil {
		return nil, fmt.Errorf("read a policy file: %w", err)
	}
	cfg.Allowed = true
	reader.set(policyFilePath, cfg)
	return cfg, nil
}
