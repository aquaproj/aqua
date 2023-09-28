package policy

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Reader interface {
	Read(policyFilePaths []string) ([]*Config, error)
	Append(logE *logrus.Entry, aquaYAMLPath string, policies []*Config, globalPolicyPaths map[string]struct{}) ([]*Config, error)
}

type MockReader struct {
	Config  *Config
	Configs []*Config
	Err     error
}

func (r *MockReader) Read(policyFilePaths []string) ([]*Config, error) {
	allowCfgs(r.Configs)
	return r.Configs, r.Err
}

func (r *MockReader) Append(logE *logrus.Entry, aquaYAMLPath string, policies []*Config, globalPolicyPaths map[string]struct{}) ([]*Config, error) {
	return r.Configs, r.Err
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

func (r *ReaderImpl) Read(policyFilePaths []string) ([]*Config, error) {
	cfgs, err := r.reader.Read(policyFilePaths)
	if err != nil {
		return nil, fmt.Errorf("read policies from the environment variable: %w", err)
	}
	allowCfgs(cfgs)
	return cfgs, nil
}

func (r *ReaderImpl) Append(logE *logrus.Entry, aquaYAMLPath string, policies []*Config, globalPolicyPaths map[string]struct{}) ([]*Config, error) {
	policyFilePath, err := r.finder.Find("", filepath.Dir(aquaYAMLPath))
	if err != nil {
		return nil, fmt.Errorf("find a policy file: %w", err)
	}
	if policyFilePath == "" {
		return policies, nil
	}
	if _, ok := globalPolicyPaths[policyFilePath]; ok {
		return policies, nil
	}
	policyCfg, err := r.read(logE, policyFilePath)
	if err != nil {
		return nil, fmt.Errorf("read a policy file: %w", err)
	}
	if policyCfg == nil {
		return policies, nil
	}
	return append(policies, policyCfg), nil
}

func (r *ReaderImpl) get(p string) *Config {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.policies[p]
}

func (r *ReaderImpl) set(p string, cfg *Config) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.policies[p] = cfg
}

func allowCfgs(cfgs []*Config) {
	for _, cfg := range cfgs {
		cfg.Allowed = true
	}
}

func (r *ReaderImpl) read(logE *logrus.Entry, policyFilePath string) (*Config, error) {
	if cfg := r.get(policyFilePath); cfg != nil {
		if cfg.Allowed {
			return cfg, nil
		}
		return nil, nil //nolint:nilnil
	}
	if err := r.validator.Validate(policyFilePath); err != nil {
		r.set(policyFilePath, &Config{})
		if err := r.validator.Warn(logE, policyFilePath, errors.Is(err, errPolicyUpdated)); err != nil {
			logE.WithError(err).Warn("warn an denied policy file")
		}
		return nil, nil //nolint:nilnil
	}
	cfg, err := r.reader.ReadFile(policyFilePath)
	if err != nil {
		return nil, fmt.Errorf("read a policy file: %w", err)
	}
	cfg.Allowed = true
	r.set(policyFilePath, cfg)
	return cfg, nil
}
