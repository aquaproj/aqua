package reader

// import (
// 	"fmt"
// 	"path/filepath"
// 	"sort"
// 	"strings"
//
// 	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
// 	"github.com/aquaproj/aqua/v2/pkg/util"
// 	"github.com/spf13/afero"
// 	"gopkg.in/yaml.v2"
// )
//
// func (reader *ConfigReaderImpl) ReadToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error) {
// 	file, err := reader.fs.Open(configFilePath)
// 	if err != nil {
// 		return nil, err //nolint:wrapcheck
// 	}
// 	defer file.Close()
// 	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
// 		return nil, fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
// 	}
// 	var configFileDir string
// 	for _, rgst := range cfg.Registries {
// 		rgst := rgst
// 		if rgst.Type == "local" {
// 			if strings.HasPrefix(rgst.Path, homePrefix) {
// 				if reader.homeDir == "" {
// 					return nil, errHomeDirEmpty
// 				}
// 				rgst.Path = filepath.Join(reader.homeDir, rgst.Path[6:])
// 			}
// 			if configFileDir == "" {
// 				configFileDir = filepath.Dir(configFilePath)
// 			}
// 			rgst.Path = util.Abs(configFileDir, rgst.Path)
// 		}
// 	}
// 	cfgs, err := reader.readImportsToUpdate(configFilePath, cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("read imports (%s): %w", configFilePath, err)
// 	}
// 	return cfgs, nil
// }
//
// func (reader *ConfigReaderImpl) readImportsToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error) {
// 	cfgs := map[string]*aqua.Config{}
// 	pkgs := []*aqua.Package{}
// 	for _, pkg := range cfg.Packages {
// 		if pkg == nil {
// 			continue
// 		}
// 		if pkg.Import == "" {
// 			pkgs = append(pkgs, pkg)
// 			continue
// 		}
// 		p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
// 		filePaths, err := afero.Glob(reader.fs, p)
// 		if err != nil {
// 			return nil, fmt.Errorf("read files with glob pattern (%s): %w", p, err)
// 		}
// 		sort.Strings(filePaths)
// 		for _, filePath := range filePaths {
// 			subCfg := &aqua.Config{}
// 			if err := reader.Read(filePath, subCfg); err != nil {
// 				return nil, err
// 			}
// 			cfgs[filePath] = subCfg
// 		}
// 	}
// 	cfg.Packages = pkgs
// 	return cfgs, nil
// }
