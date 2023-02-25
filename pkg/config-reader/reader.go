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
	"github.com/aquaproj/aqua/pkg/yaml"
	"github.com/spf13/afero"
)

var errHomeDirEmpty = errors.New("failed to get a user home directory")

type ConfigReaderImpl struct {
	fs          afero.Fs
	homeDir     string
	yamlDecoder *yaml.Decoder
}

func New(fs afero.Fs, param *config.Param) *ConfigReaderImpl {
	return &ConfigReaderImpl{
		fs:          fs,
		homeDir:     param.HomeDir,
		yamlDecoder: yaml.NewDecoder(fs),
	}
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type MockConfigReader struct {
	Cfg *aqua.Config
	Err error
}

func (reader *MockConfigReader) Read(configFilePath string, cfg *aqua.Config) error {
	*cfg = *reader.Cfg
	return reader.Err
}

const homePrefix = "$HOME" + string(os.PathSeparator)

func (reader *ConfigReaderImpl) Read(configFilePath string, cfg *aqua.Config) error {
	if err := reader.yamlDecoder.ReadFile(configFilePath, cfg); err != nil {
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

func (reader *ConfigReaderImpl) readImports(configFilePath string, cfg *aqua.Config) error {
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
