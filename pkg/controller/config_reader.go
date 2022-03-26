package controller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"
)

type FileReader interface {
	Read(p string) (io.ReadCloser, error)
}

type fileReader struct{}

func (reader *fileReader) Read(p string) (io.ReadCloser, error) {
	return os.Open(p) //nolint:wrapcheck
}

type configReader struct {
	reader FileReader
}

func (reader *configReader) Read(configFilePath string, cfg *Config) error {
	file, err := reader.reader.Read(configFilePath)
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

func (reader *configReader) readImports(configFilePath string, cfg *Config) error {
	pkgs := []*Package{}
	for _, pkg := range cfg.Packages {
		if pkg.Import == "" {
			pkgs = append(pkgs, pkg)
			continue
		}
		p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
		filePaths, err := filepath.Glob(p)
		if err != nil {
			return fmt.Errorf("read files with glob pattern (%s): %w", p, err)
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			subCfg := &Config{}
			if err := reader.Read(filePath, subCfg); err != nil {
				return err
			}
			pkgs = append(pkgs, subCfg.Packages...)
		}
	}
	cfg.Packages = pkgs
	return nil
}
