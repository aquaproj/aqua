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
	r.readImports(logE, configFilePath, cfg)
	return nil
}

func (r *ConfigReader) readImports(logE *logrus.Entry, configFilePath string, cfg *aqua.Config) {
	pkgs := []*aqua.Package{}
	for _, pkg := range cfg.Packages {
		if pkg == nil {
			continue
		}
		if pkg.Import == "" {
			if err := readGoVersionFile(r.fs, configFilePath, pkg); err != nil {
				logerr.WithError(logE, err).Error("read a go version file")
				continue
			}
			pkgs = append(pkgs, pkg)
			continue
		}
		logE := logE.WithField("import", pkg.Import)
		p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
		filePaths, err := afero.Glob(r.fs, p)
		if err != nil {
			logerr.WithError(logE, err).Error("read files with glob pattern")
			continue
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			logE := logE.WithField("imported_file", filePath)
			subCfg := &aqua.Config{}
			if err := r.Read(logE, filePath, subCfg); err != nil {
				logerr.WithError(logE, err).Error("read an import file")
				continue
			}
			for _, pkg := range subCfg.Packages {
				pkg.FilePath = filePath
				if err := readGoVersionFile(r.fs, filePath, pkg); err != nil {
					logerr.WithError(logE, err).Error("read a go version file")
					continue
				}
				pkgs = append(pkgs, pkg)
			}
		}
	}
	cfg.Packages = pkgs
}
