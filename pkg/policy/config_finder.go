package policy

import (
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
)

type ConfigFinder interface {
	Find(policyFilePath, wd string) (string, error)
}

type ConfigFinderImpl struct{}

func NewConfigFinder() *ConfigFinderImpl {
	return &ConfigFinderImpl{}
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
		if !osfile.Exists(policyFilePath) {
			return "", ErrConfigFileNotFound
		}
		if filepath.IsAbs(policyFilePath) {
			return policyFilePath, nil
		}
		return filepath.Join(wd, policyFilePath), nil
	}

	// https://github.com/orgs/aquaproj/discussions/2476
	// Using `git worktree`, .git is a file.
	gitDir := findconfig.Find(wd, osfile.Exists, ".git")
	if gitDir == "" {
		return "", nil
	}
	gitParentDir := filepath.Dir(gitDir)
	for _, p := range configFileNames() {
		if p := filepath.Join(gitParentDir, p); osfile.Exists(p) {
			return p, nil
		}
	}
	return "", nil
}
