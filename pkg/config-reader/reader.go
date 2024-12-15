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
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

var errHomeDirEmpty = errors.New("failed to get a user home directory")

type ConfigReader struct {
	fs      afero.Fs
	homeDir string
}

func New(fs afero.Fs, param *config.Param) *ConfigReader {
	return &ConfigReader{
		fs:      fs,
		homeDir: param.HomeDir,
	}
}

const homePrefix = "$HOME" + string(os.PathSeparator)

func (r *ConfigReader) Read(logE *logrus.Entry, configFilePath string, cfg *aqua.Config) error {
	logE = logE.WithField("config_file_path", configFilePath)
	file, err := r.fs.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML: %w", err)
	}
	configFileDir := filepath.Dir(configFilePath)
	if err := r.readRegistries(configFileDir, cfg); err != nil {
		return err
	}
	r.readPackages(logE, configFilePath, cfg)
	return nil
}

func (r *ConfigReader) readRegistries(configFileDir string, cfg *aqua.Config) error {
	for _, rgst := range cfg.Registries {
		if rgst.Type == "local" {
			if strings.HasPrefix(rgst.Path, homePrefix) {
				if r.homeDir == "" {
					return errHomeDirEmpty
				}
				rgst.Path = filepath.Join(r.homeDir, rgst.Path[6:])
			}
			rgst.Path = osfile.Abs(configFileDir, rgst.Path)
		}
	}
	return nil
}

func (r *ConfigReader) readPackages(logE *logrus.Entry, configFilePath string, cfg *aqua.Config) {
	pkgs := []*aqua.Package{}
	for _, pkg := range cfg.Packages {
		if pkg == nil {
			continue
		}
		subPkgs, err := r.readPackage(logE, configFilePath, pkg)
		if err != nil {
			logerr.WithError(logE, err).Error("read a package")
			continue
		}
		if subPkgs == nil {
			pkg.FilePath = configFilePath
			pkgs = append(pkgs, pkg)
			continue
		}
		pkgs = append(pkgs, subPkgs...)
	}
	cfg.Packages = pkgs
}

func (r *ConfigReader) readPackage(logE *logrus.Entry, configFilePath string, pkg *aqua.Package) ([]*aqua.Package, error) {
	if pkg.GoVersionFile != "" {
		// go_version_file
		if err := readGoVersionFile(r.fs, configFilePath, pkg); err != nil {
			return nil, fmt.Errorf("read a go version file: %w", logerr.WithFields(err, logrus.Fields{
				"go_version_file": pkg.GoVersionFile,
			}))
		}
		return nil, nil
	}
	if pkg.Import == "" {
		// version
		return nil, nil
	}
	// import
	logE = logE.WithField("import", pkg.Import)
	p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
	filePaths, err := afero.Glob(r.fs, p)
	if err != nil {
		return nil, fmt.Errorf("find files with a glob pattern: %w", err)
	}
	sort.Strings(filePaths)
	pkgs := []*aqua.Package{}
	for _, filePath := range filePaths {
		logE := logE.WithField("imported_file", filePath)
		subCfg := &aqua.Config{}
		if err := r.Read(logE, filePath, subCfg); err != nil {
			logerr.WithError(logE, err).Error("read an import file")
			continue
		}
		pkgs = append(pkgs, subCfg.Packages...)
	}
	return pkgs, nil
}
