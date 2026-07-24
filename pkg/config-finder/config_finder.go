package finder

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
)

var ErrConfigFileNotFound = errors.New("configuration file isn't found")

type ConfigFinder struct{}

func NewConfigFinder() *ConfigFinder {
	return &ConfigFinder{}
}

func ParseGlobalConfigFilePaths(pwd, env string) []string {
	src := filepath.SplitList(env)
	paths := make([]string, 0, len(src))
	m := make(map[string]struct{}, len(src))
	for _, s := range src {
		if s == "" {
			continue
		}
		s = filepath.Clean(s)
		if !filepath.IsAbs(s) {
			s = filepath.Join(pwd, s)
		}
		if _, ok := m[s]; ok {
			continue
		}
		m[s] = struct{}{}
		paths = append(paths, s)
	}
	return paths
}

func ConfigFileNames() []string {
	return []string{
		"aqua.yaml",
		"aqua.yml",
		".aqua.yaml",
		".aqua.yml",
		filepath.Join("aqua", "aqua.yaml"),
		filepath.Join("aqua", "aqua.yml"),
		filepath.Join(".aqua", "aqua.yaml"),
		filepath.Join(".aqua", "aqua.yml"),
	}
}

func DuplicateFilePaths(filePath string) []string {
	filePaths := ConfigFileNames()
	fileNames := map[string]struct{}{}
	for _, p := range filePaths {
		fileNames[filepath.Base(p)] = struct{}{}
	}
	fileName := filepath.Base(filePath)
	if _, ok := fileNames[fileName]; !ok {
		return nil
	}
	dir := filepath.Dir(filePath)
	parentDir := filepath.Base(dir)
	paths := []string{}
	if (parentDir == "aqua" || parentDir == ".aqua") && !strings.HasPrefix(fileName, ".") {
		// e.g. aqua/aqua.yaml
		ddir := filepath.Dir(dir)
		for _, p := range filePaths {
			paths = append(paths, filepath.Join(ddir, p))
		}
		return paths
	}
	for _, p := range filePaths {
		paths = append(paths, filepath.Join(dir, p))
	}
	return paths
}

func (f *ConfigFinder) Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error) {
	if configFilePath != "" {
		return osfile.Abs(wd, configFilePath), nil
	}
	configFilePath = findconfig.Find(wd, exists, ConfigFileNames()...)
	if configFilePath != "" {
		return configFilePath, nil
	}
	for _, p := range globalConfigFilePaths {
		if f, err := osfile.Exists(p); err != nil {
			return "", err //nolint:wrapcheck
		} else if f {
			return p, nil
		}
	}
	return "", ErrConfigFileNotFound
}

func (f *ConfigFinder) Finds(wd, configFilePath string) []string {
	if configFilePath == "" {
		return findconfig.Finds(wd, exists, ConfigFileNames()...)
	}
	return []string{osfile.Abs(wd, configFilePath)}
}

// exists adapts osfile.Exists to findconfig's predicate, which walks up the
// directory tree and has no way to report an error. A path that can't be
// stat'd is treated as absent, so the walk moves on to the parent directory.
func exists(p string) bool {
	f, _ := osfile.Exists(p)
	return f
}
