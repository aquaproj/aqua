package reader

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func New(fs afero.Fs) *ConfigReader {
	return &ConfigReader{
		fs: fs,
	}
}

type ConfigReader struct {
	fs afero.Fs
}

func (reader *ConfigReader) Read(configFilePath string, cfg *aqua.Config) error {
	file, err := reader.fs.Open(configFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
	}
	if err := reader.readImports(configFilePath, cfg); err != nil {
		return fmt.Errorf("read imports (%s): %w", configFilePath, err)
	}
	return nil
}

func (reader *ConfigReader) readImports(configFilePath string, cfg *aqua.Config) error {
	pkgs := []*aqua.Package{}
	for _, pkg := range cfg.Packages {
		if pkg.Import == "" {
			pkgs = append(pkgs, pkg)
			continue
		}
		p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
		filePaths, err := afero.Glob(reader.fs, p)
		if err != nil {
			return fmt.Errorf("read files with glob pattern (%s): %w", p, err)
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			subCfg := &aqua.Config{}
			if err := reader.Read(filePath, subCfg); err != nil {
				return err
			}
			pkgs = append(pkgs, subCfg.Packages...)
		}
	}
	cfg.Packages = pkgs
	return nil
}
