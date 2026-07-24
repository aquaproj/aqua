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
		// The user named this file, so a stat failure other than "not found"
		// must not be reported as "not found".
		if e, err := osfile.Exists(policyFilePath); err != nil {
			return "", err //nolint:wrapcheck
		} else if !e {
			return "", ErrConfigFileNotFound
		}
		if filepath.IsAbs(policyFilePath) {
			return policyFilePath, nil
		}
		return filepath.Join(wd, policyFilePath), nil
	}

	// https://github.com/orgs/aquaproj/discussions/2476
	// Using `git worktree`, .git is a file.
	gitDir := findconfig.Find(wd, existsBestEffort, ".git")
	if gitDir == "" {
		return "", nil
	}
	gitParentDir := filepath.Dir(gitDir)
	for _, p := range configFileNames() {
		p := filepath.Join(gitParentDir, p)
		if e, err := osfile.Exists(p); err != nil {
			return "", err //nolint:wrapcheck
		} else if e {
			return p, nil
		}
	}
	return "", nil
}

// exists adapts osfile.Exists to findconfig's predicate, which walks up the
// directory tree and has no way to report an error. A path that can't be
// stat'd is treated as absent, so the walk moves on to the parent directory.
func existsBestEffort(p string) bool {
	f, _ := osfile.Exists(p)
	return f
}
