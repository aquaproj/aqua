package finder

import (
	"errors"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
)

var ErrConfigFileNotFound = errors.New("configuration file isn't found")

type ConfigFinder struct {
	fs afero.Fs
}

func NewConfigFinder(fs afero.Fs) *ConfigFinder {
	return &ConfigFinder{
		fs: fs,
	}
}

func ParseGlobalConfigFilePaths(env string) []string {
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

func configFileNames() []string {
	return []string{"aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml"}
}

func (finder *ConfigFinder) Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error) {
	if configFilePath != "" {
		return configFilePath, nil
	}
	configFilePath = findconfig.Find(wd, finder.exist, configFileNames()...)
	if configFilePath != "" {
		return configFilePath, nil
	}
	for _, p := range globalConfigFilePaths {
		if _, err := finder.fs.Stat(p); err != nil {
			continue
		}
		return p, nil
	}
	return "", ErrConfigFileNotFound
}

func (finder *ConfigFinder) Finds(wd, configFilePath string) []string {
	if configFilePath == "" {
		return findconfig.Finds(wd, finder.exist, configFileNames()...)
	}
	return append([]string{configFilePath}, findconfig.Finds(wd, finder.exist, configFileNames()...)...)
}

func (finder *ConfigFinder) exist(p string) bool {
	b, err := afero.Exists(finder.fs, p)
	if err != nil {
		return false
	}
	return b
}
