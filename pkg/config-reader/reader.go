package reader

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/clivm/clivm/pkg/config/aqua"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type ConfigReader interface {
	Read(configFilePath string, cfg *clivm.Config) error
}

func New(fs afero.Fs) ConfigReader {
	return &configReader{
		fs: fs,
	}
}

type configReader struct {
	fs afero.Fs
}

func (reader *configReader) Read(configFilePath string, cfg *clivm.Config) error {
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

func (reader *configReader) readImports(configFilePath string, cfg *clivm.Config) error {
	pkgs := []*clivm.Package{}
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
			subCfg := &clivm.Config{}
			if err := reader.Read(filePath, subCfg); err != nil {
				return err
			}
			pkgs = append(pkgs, subCfg.Packages...)
		}
	}
	cfg.Packages = pkgs
	return nil
}
