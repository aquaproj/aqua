package reader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

var errHomeDirEmpty = errors.New("failed to get a user home directory")

func New(fs afero.Fs, param *config.Param) *ConfigReader {
	return &ConfigReader{
		fs:      fs,
		homeDir: param.HomeDir,
	}
}

type ConfigReader struct {
	fs      afero.Fs
	homeDir string
}

const homePrefix = "$HOME" + string(os.PathSeparator)

func (reader *ConfigReader) Read(configFilePath string, cfg *aqua.Config) error {
	file, err := reader.fs.Open(configFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
	}
	var configFileDir string
	for _, rgst := range cfg.Registries {
		rgst := rgst
		if rgst.Type == "local" {
			if strings.HasPrefix(rgst.Path, homePrefix) {
				if reader.homeDir == "" {
					return errHomeDirEmpty
				}
				rgst.Path = filepath.Join(reader.homeDir, rgst.Path[6:])
			}
			if configFileDir == "" {
				configFileDir = filepath.Dir(configFilePath)
			}
			rgst.Path = util.Abs(configFileDir, rgst.Path)
		}
	}
	if err := reader.readImports(configFilePath, cfg); err != nil {
		return fmt.Errorf("read imports (%s): %w", configFilePath, err)
	}
	return nil
}

func (reader *ConfigReader) readImports(configFilePath string, cfg *aqua.Config) error {
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
