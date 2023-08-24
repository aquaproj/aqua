package policy

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
)

type ConfigFinder interface {
	Find(policyFilePath, wd string) (string, error)
}

type MockConfigFinder struct {
	path string
	err  error
}

func (f *MockConfigFinder) Find(policyFilePath, wd string) (string, error) {
	return f.path, f.err
}

type ConfigFinderImpl struct {
	fs afero.Fs
}

func NewConfigFinder(fs afero.Fs) *ConfigFinderImpl {
	return &ConfigFinderImpl{
		fs: fs,
	}
}

func configFileNames() []string {
	return []string{
		"aqua-policy.yaml",
		".aqua-policy.yaml",
		filepath.Join("aqua", "aqua-policy.yaml"),
		filepath.Join(".aqua", "aqua-policy.yaml"),
	}
}

func (f *ConfigFinderImpl) Find(policyFilePath, wd string) (string, error) {
	if policyFilePath != "" {
		f, err := afero.Exists(f.fs, policyFilePath)
		if err != nil {
			return "", fmt.Errorf("check if a policy file exists: %w", err)
		}
		if !f {
			return "", ErrConfigFileNotFound
		}
		if filepath.IsAbs(policyFilePath) {
			return policyFilePath, nil
		}
		return filepath.Join(wd, policyFilePath), nil
	}

	gitDir := findconfig.Find(wd, f.exist, ".git")
	if gitDir == "" {
		return "", nil
	}
	gitParentDir := filepath.Dir(gitDir)
	for _, p := range configFileNames() {
		if _, err := f.fs.Stat(filepath.Join(gitParentDir, p)); err != nil {
			continue
		}
		return filepath.Join(gitParentDir, p), nil
	}
	return "", nil
}

func (f *ConfigFinderImpl) exist(p string) bool {
	b, err := afero.IsDir(f.fs, p)
	if err != nil {
		return false
	}
	return b
}
