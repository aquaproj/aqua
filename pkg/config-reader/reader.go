package reader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

var errHomeDirEmpty = errors.New("failed to get a user home directory")

type ConfigReaderImpl struct {
	fs      afero.Fs
	homeDir string
}

func New(fs afero.Fs, param *config.Param) *ConfigReaderImpl {
	return &ConfigReaderImpl{
		fs:      fs,
		homeDir: param.HomeDir,
	}
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type MockConfigReader struct {
	Cfg *aqua.Config
	Err error
}

func (r *MockConfigReader) Read(configFilePath string, cfg *aqua.Config) error {
	*cfg = *r.Cfg
	return r.Err
}

const homePrefix = "$HOME" + string(os.PathSeparator)

func (r *ConfigReaderImpl) Read(configFilePath string, cfg *aqua.Config) error {
	file, err := r.fs.Open(configFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
	}
	var configFileDir string
	for _, rgst := range cfg.Registries {
		if rgst.Type == "local" {
			if strings.HasPrefix(rgst.Path, homePrefix) {
				if r.homeDir == "" {
					return errHomeDirEmpty
				}
				rgst.Path = filepath.Join(r.homeDir, rgst.Path[6:])
			}
			if configFileDir == "" {
				configFileDir = filepath.Dir(configFilePath)
			}
			rgst.Path = osfile.Abs(configFileDir, rgst.Path)
		}
	}
	if err := r.readImports(configFilePath, cfg); err != nil {
		return fmt.Errorf("read imports (%s): %w", configFilePath, err)
	}
	return nil
}

func (r *ConfigReaderImpl) readImports(configFilePath string, cfg *aqua.Config) error {
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
			return fmt.Errorf("read files with glob pattern (%s): %w", p, err)
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			subCfg := &aqua.Config{}
			if err := r.Read(filePath, subCfg); err != nil {
				return err
			}
			for _, pkg := range subCfg.Packages {
				pkg.FilePath = filePath
				pkgs = append(pkgs, pkg)
			}
		}
	}
	cfg.Packages = pkgs
	return nil
}
