package reader

import (
	"fmt"
	"maps"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v2"
)

func (r *ConfigReader) ReadToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error) {
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

func (r *ConfigReader) readImportsToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error) { //nolint:cyclop
	cfgs := map[string]*aqua.Config{}
	pkgs := []*aqua.Package{}
	for _, pkg := range cfg.Packages {
		if pkg == nil {
			continue
		}
		if pkg.VersionExpr != "" || pkg.GoVersionFile != "" || pkg.Pin {
			// Exclude them from the update targets
			continue
		}
		if pkg.Import == "" {
			pkgs = append(pkgs, pkg)
			continue
		}
		if err := r.readImportToUpdate(configFilePath, pkg.Import, cfg, cfgs); err != nil {
			return nil, err
		}
	}
	if cfg.ImportDir != "" {
		if err := r.readImportToUpdate(configFilePath, filepath.Join(cfg.ImportDir, "*.yml"), cfg, cfgs); err != nil {
			return nil, err
		}
		if err := r.readImportToUpdate(configFilePath, filepath.Join(cfg.ImportDir, "*.yaml"), cfg, cfgs); err != nil {
			return nil, err
		}
	}
	cfg.Packages = pkgs
	return cfgs, nil
}

func (r *ConfigReader) readImportToUpdate(configFilePath, importPath string, cfg *aqua.Config, cfgs map[string]*aqua.Config) error {
	p := filepath.Join(filepath.Dir(configFilePath), importPath)
	filePaths, err := afero.Glob(r.fs, p)
	if err != nil {
		return fmt.Errorf("read files with glob pattern (%s): %w", p, err)
	}
	sort.Strings(filePaths)
	for _, filePath := range filePaths {
		subCfg := &aqua.Config{}
		subCfgs, err := r.ReadToUpdate(filePath, subCfg)
		if err != nil {
			return err
		}
		subCfg.Registries = cfg.Registries
		cfgs[filePath] = subCfg
		maps.Copy(cfgs, subCfgs)
	}
	return nil
}
