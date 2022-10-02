package updatechecksum

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	rootDir           string
	configFinder      ConfigFinder
	configReader      domain.ConfigReader
	registryInstaller domain.RegistryInstaller
	fs                afero.Fs
	runtime           *runtime.Runtime
	chkDL             domain.ChecksumDownloader
	parser            *checksum.FileParser
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

func New(param *config.Param, configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller, fs afero.Fs, rt *runtime.Runtime, chkDL domain.ChecksumDownloader) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		fs:                fs,
		runtime:           rt,
		chkDL:             chkDL,
		parser:            &checksum.FileParser{},
	}
}

func (ctrl *Controller) UpdateChecksum(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := ctrl.updateChecksum(ctx, logE, cfgFilePath); err != nil {
			return err
		}
	}

	return ctrl.updateChecksumAll(ctx, logE, param)
}

func (ctrl *Controller) updateChecksumAll(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.updateChecksum(ctx, logE, cfgFilePath); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) updateChecksum(ctx context.Context, logE *logrus.Entry, cfgFilePath string) (namedErr error) {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}

	checksums := checksum.New()
	checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(ctrl.fs, cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if err := checksums.ReadFile(ctrl.fs, checksumFilePath); err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	pkgs, _ := config.ListPackagesNotOverride(logE, cfg, registryContents)
	failed := false
	defer func() {
		if err := checksums.UpdateFile(ctrl.fs, checksumFilePath); err != nil {
			namedErr = fmt.Errorf("update a checksum file: %w", err)
		}
	}()
	for _, pkg := range pkgs {
		logE := logE.WithFields(logrus.Fields{
			"package_name":     pkg.Package.Name,
			"package_version":  pkg.Package.Version,
			"package_registry": pkg.Package.Registry,
		})
		logE.Info("updating a package checksum")
		if err := ctrl.updatePackage(ctx, logE, checksums, pkg); err != nil {
			failed = true
			logerr.WithError(logE, err).Error("update checksums")
		}
	}
	if failed {
		return errFailedToUpdateChecksum
	}
	return nil
}

func (ctrl *Controller) updatePackage(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, pkg *config.Package) error {
	if err := ctrl.getChecksums(ctx, logE, checksums, pkg); err != nil {
		return err
	}
	return nil
}

func (ctrl *Controller) getChecksums(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, pkg *config.Package) error { //nolint:funlen,cyclop
	rts, err := runtime.GetRuntimesFromEnvs(pkg.PackageInfo.SupportedEnvs)
	if err != nil {
		return fmt.Errorf("get supported platforms: %w", err)
	}
	checksumFiles := map[string]struct{}{}
	for _, rt := range rts {
		rt := rt
		logE := logE.WithFields(logrus.Fields{
			"checksum_env": rt.GOOS + "/" + rt.GOARCH,
		})
		pkgInfo := pkg.PackageInfo
		pkgInfo = pkgInfo.Copy()
		pkgInfo.OverrideByRuntime(rt)
		pkg := &config.Package{
			Package:     pkg.Package,
			PackageInfo: pkgInfo,
		}

		if !pkg.PackageInfo.Checksum.GetEnabled() {
			logE.Debug("chekcsum isn't supported")
			continue
		}

		checksumID, err := pkg.GetChecksumID(rt)
		if err != nil {
			return fmt.Errorf("get a checksum id: %w", err)
		}

		if a := checksums.Get(checksumID); a != nil {
			continue
		}

		checksumFileID, err := pkg.RenderChecksumFileID(rt)
		if err != nil {
			return fmt.Errorf("render a checksum file ID: %w", err)
		}
		if _, ok := checksumFiles[checksumFileID]; ok {
			continue
		}
		checksumFiles[checksumFileID] = struct{}{}
		logE.Debug("downloading a checksum file")
		file, _, err := ctrl.chkDL.DownloadChecksum(ctx, logE, rt, pkg)
		if err != nil {
			return fmt.Errorf("download a checksum file: %w", err)
		}
		if file == nil {
			continue
		}
		defer file.Close()
		b, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("read a checksum file: %w", err)
		}
		checksumFile := strings.TrimSpace(string(b))
		if pkgInfo.Checksum.FileFormat == "raw" {
			logE.WithFields(logrus.Fields{
				"checksum_id": checksumID,
				"checksum":    checksumFile,
			}).Debug("set a checksum")
			checksums.Set(checksumID, &checksum.Checksum{
				ID:        checksumID,
				Checksum:  checksumFile,
				Algorithm: pkgInfo.Checksum.GetAlgorithm(),
			})
			continue
		}
		m, err := ctrl.parser.ParseChecksumFile(checksumFile, pkg)
		if err != nil {
			return fmt.Errorf("parse a checksum file: %w", err)
		}
		for assetName, chksum := range m {
			checksumID, err := pkg.GetChecksumIDFromAsset(assetName)
			if err != nil {
				return fmt.Errorf("get a checksum id from asset: %w", err)
			}
			logE.WithFields(logrus.Fields{
				"checksum_id": checksumID,
				"checksum":    chksum,
			}).Debug("set a checksum")
			checksums.Set(checksumID, &checksum.Checksum{
				ID:        checksumID,
				Checksum:  chksum,
				Algorithm: pkgInfo.Checksum.GetAlgorithm(),
			})
		}
	}
	return nil
}
