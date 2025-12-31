package policy

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type Reader struct {
	mutex     *sync.RWMutex
	policies  map[string]*Config
	fs        afero.Fs
	validator Validator
	finder    ConfigFinder
	reader    ConfigReader
}

func NewReader(fs afero.Fs, validator Validator, finder ConfigFinder, reader ConfigReader) *Reader {
	return &Reader{
		mutex:     &sync.RWMutex{},
		policies:  map[string]*Config{},
		fs:        fs,
		validator: validator,
		finder:    finder,
		reader:    reader,
	}
}

func (r *Reader) Read(policyFilePaths []string) ([]*Config, error) {
	cfgs, err := r.reader.Read(policyFilePaths)
	if err != nil {
		return nil, fmt.Errorf("read policies from the environment variable: %w", err)
	}
	allowCfgs(cfgs)
	return cfgs, nil
}

// Append finds and reads a policy file for aquaYAMLPath and appends the policy to policies.
func (r *Reader) Append(logger *slog.Logger, aquaYAMLPath string, policies []*Config, globalPolicyPaths map[string]struct{}) ([]*Config, error) {
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
	policyCfg, err := r.read(logger, policyFilePath)
	if err != nil {
		return nil, fmt.Errorf("read a policy file: %w", err)
	}
	if policyCfg == nil {
		return policies, nil
	}
	return append(policies, policyCfg), nil
}

func (r *Reader) get(p string) *Config {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.policies[p]
}

func (r *Reader) set(p string, cfg *Config) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.policies[p] = cfg
}

func allowCfgs(cfgs []*Config) {
	for _, cfg := range cfgs {
		cfg.Allowed = true
	}
}

func (r *Reader) read(logger *slog.Logger, policyFilePath string) (*Config, error) {
	if cfg := r.get(policyFilePath); cfg != nil {
		if cfg.Allowed {
			return cfg, nil
		}
		return nil, nil //nolint:nilnil
	}
	if err := r.validator.Validate(policyFilePath); err != nil {
		r.set(policyFilePath, &Config{})
		if err := r.validator.Warn(logger, policyFilePath, errors.Is(err, errPolicyUpdated)); err != nil {
			slogerr.WithError(logger, err).Warn("warn a denied policy file")
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
