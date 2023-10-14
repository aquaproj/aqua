package reader

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func (r *ConfigReaderImpl) ReadToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error) {
	file, err := r.fs.Open(configFilePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
	}
	var configFileDir string
	for _, rgst := range cfg.Registries {
		rgst := rgst
		if rgst.Type == "local" {
			if strings.HasPrefix(rgst.Path, homePrefix) {
				if r.homeDir == "" {
					return nil, errHomeDirEmpty
				}
				rgst.Path = filepath.Join(r.homeDir, rgst.Path[6:]) // 6: "$HOME/"
			}
			if configFileDir == "" {
				configFileDir = filepath.Dir(configFilePath)
			}
			rgst.Path = osfile.Abs(configFileDir, rgst.Path)
		}
	}
	cfgs, err := r.readImportsToUpdate(configFilePath, cfg)
	if err != nil {
		return nil, fmt.Errorf("read imports (%s): %w", configFilePath, err)
	}
	return cfgs, nil
}

func (r *ConfigReaderImpl) readImportsToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error) {
	cfgs := map[string]*aqua.Config{}
	pkgs := []*aqua.Package{}
	for _, pkg := range cfg.Packages {
		if pkg == nil {
			continue
		}
		if pkg.Import == "" {
			pkgs = append(pkgs, pkg)
			continue
		}
		p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
		filePaths, err := afero.Glob(r.fs, p)
		if err != nil {
			return nil, fmt.Errorf("read files with glob pattern (%s): %w", p, err)
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			subCfg := &aqua.Config{}
			if err := r.Read(filePath, subCfg); err != nil {
				return nil, err
			}
			subCfg.Registries = cfg.Registries
			cfgs[filePath] = subCfg
		}
	}
	cfg.Packages = pkgs
	return cfgs, nil
}
