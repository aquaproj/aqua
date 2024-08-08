package packageinfo

import (
	"errors"
	"encoding/json"
	"strings"
	"fmt"
	"io"
	"os"
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/spf13/afero"
	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
)

var (
	errMissingPackages = errors.New("could not find all packages")
)

type Controller struct {
	stdout            io.Writer
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	fs                afero.Fs
}

type ConfigReader interface {
	Read(logE *logrus.Entry, configFilePath string, cfg *aqua.Config) error
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
	Finds(wd, configFilePath string) []string
}

func New(configFinder ConfigFinder, configReader ConfigReader, registInstaller RegistryInstaller, fs afero.Fs) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		fs:                fs,
	}
}

func (c *Controller) PackageInfo(ctx context.Context, param *config.Param, logE *logrus.Entry, args ...string) error {
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err
	}

	cfg := &aqua.Config{}
	if err := c.configReader.Read(logE, cfgFilePath, cfg); err != nil {
		return err
	}

	var checksums *checksum.Checksums
	if cfg.ChecksumEnabled(param.EnforceChecksum, param.Checksum) {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, cfgFilePath)
		if err != nil {
			return err
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err
	}

	packageInfoByNameByRegistry := make(map[string]map[string]*registry.PackageInfo)
	for registryName, registryConfig := range registryContents {
		packageInfoByNameByRegistry[registryName]= registryConfig.PackageInfos.ToMap(logE)
	}

	encoder := json.NewEncoder(c.stdout)
	encoder.SetIndent("", "  ")

	missingPackages := false
	for _, name := range args {
		registryName, pkgName := registryAndPackage(name)
		pkgInfo, ok := packageInfoByNameByRegistry[registryName][pkgName]
		if !ok {
			logE.WithFields(logrus.Fields {
				"registry_name": registryName,
				"package_name": pkgName,
			}).Error("no such registry or package")
			missingPackages = true
			continue
		}

		if err := encoder.Encode(pkgInfo); err != nil {
			return fmt.Errorf("encode info as JSON and output it to stdout: %w", err)
		}
	}

	if missingPackages {
		return errMissingPackages
	}

	return nil
}

func registryAndPackage(name string) (string, string) {
	r, p, ok := strings.Cut(name, ",")
	if ok {
		return r, p
	}
	return "standard", name
}
