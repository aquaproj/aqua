package updatechecksum

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	rootDir            string
	configFinder       ConfigFinder
	configReader       reader.ConfigReader
	registryInstaller  registry.Installer
	registryDownloader domain.GitHubContentFileDownloader
	fs                 afero.Fs
	runtime            *runtime.Runtime
	chkDL              download.ChecksumDownloader
	downloader         download.ClientAPI
	prune              bool
}

func New(param *config.Param, configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, fs afero.Fs, rt *runtime.Runtime, chkDL download.ChecksumDownloader, pkgDownloader download.ClientAPI, registDownloader domain.GitHubContentFileDownloader) *Controller {
	return &Controller{
		rootDir:            param.RootDir,
		configFinder:       configFinder,
		configReader:       configReader,
		registryInstaller:  registInstaller,
		registryDownloader: registDownloader,
		fs:                 fs,
		runtime:            rt,
		chkDL:              chkDL,
		downloader:         pkgDownloader,
		prune:              param.Prune,
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

func (ctrl *Controller) updateChecksum(ctx context.Context, logE *logrus.Entry, cfgFilePath string) (namedErr error) { //nolint:cyclop
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
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

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}
	pkgs, _ := config.ListPackagesNotOverride(logE, cfg, registryContents)
	failed := false
	defer func() {
		if ctrl.prune {
			checksums.Prune()
		}
		if err := checksums.UpdateFile(ctrl.fs, checksumFilePath); err != nil {
			namedErr = fmt.Errorf("update a checksum file: %w", err)
		}
	}()

	var supportedEnvs []string
	if cfg.Checksum != nil {
		supportedEnvs = cfg.Checksum.SupportedEnvs
	}

	for _, rgst := range cfg.Registries {
		if err := ctrl.updateRegistry(ctx, logE, checksums, rgst); err != nil {
			failed = true
			logerr.WithError(logE, err).Error("update checksums")
		}
	}

	for _, pkg := range pkgs {
		logE := logE.WithFields(logrus.Fields{
			"package_name":     pkg.Package.Name,
			"package_version":  pkg.Package.Version,
			"package_registry": pkg.Package.Registry,
		})
		if err := ctrl.updatePackage(ctx, logE, checksums, pkg, supportedEnvs); err != nil {
			failed = true
			logerr.WithError(logE, err).Error("update checksums")
		}
	}
	if failed {
		return errFailedToUpdateChecksum
	}
	return nil
}

func (ctrl *Controller) updateRegistry(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, rgst *aqua.Registry) error {
	if rgst.Type != "github_content" {
		return nil
	}
	rgstID := checksum.RegistryID(rgst)
	chksum := checksums.Get(rgstID)
	if chksum != nil {
		return nil
	}
	ghContentFile, err := ctrl.registryDownloader.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
		RepoOwner: rgst.RepoOwner,
		RepoName:  rgst.RepoName,
		Ref:       rgst.Ref,
		Path:      rgst.Path,
	})
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer ghContentFile.Close()
	content := ghContentFile.Reader()
	algorithm := "sha512"
	c, err := checksum.CalculateReader(content, "sha512")
	if err != nil {
		return fmt.Errorf("calculate a checksum of Registry: %w", err)
	}
	checksums.Set(rgstID, &checksum.Checksum{
		ID:        rgstID,
		Algorithm: algorithm,
		Checksum:  c,
	})
	return nil
}

func (ctrl *Controller) updatePackage(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error {
	if err := ctrl.getChecksums(ctx, logE, checksums, pkg, supportedEnvs); err != nil {
		return err
	}
	return nil
}

func (ctrl *Controller) getChecksums(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error {
	if pkg.PackageInfo.Type == "go_install" {
		logE.Debug("skip updating go_install package's checksum")
		return nil
	}
	logE.Info("updating a package checksum")
	rts, err := checksum.GetRuntimesFromSupportedEnvs(supportedEnvs, pkg.PackageInfo.SupportedEnvs)
	if err != nil {
		return fmt.Errorf("get supported platforms: %w", err)
	}

	pkgs, assetNames, err := ctrl.getPkgs(pkg, rts)
	if err != nil {
		return err
	}
	checksumFiles := map[string]struct{}{}
	for _, rt := range rts {
		rt := rt
		env := rt.Env()
		logE := logE.WithFields(logrus.Fields{
			"checksum_env": env,
		})
		pkg, ok := pkgs[env]
		if !ok {
			return errors.New("package isn't found")
		}
		if err := ctrl.getChecksum(ctx, logE, checksums, pkg, checksumFiles, rt, assetNames); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) getPkgs(pkg *config.Package, rts []*runtime.Runtime) (map[string]*config.Package, map[string]struct{}, error) {
	pkgs := make(map[string]*config.Package, len(rts))
	assets := make(map[string]struct{}, len(rts))
	for _, rt := range rts {
		env := rt.Env()
		pkgInfo := pkg.PackageInfo
		pkgInfo = pkgInfo.Copy()
		pkgInfo.OverrideByRuntime(rt)
		pkgWithEnv := &config.Package{
			Package:     pkg.Package,
			PackageInfo: pkgInfo,
		}
		asset, err := pkgWithEnv.RenderAsset(rt)
		if err != nil {
			return nil, nil, fmt.Errorf("render an asset: %w", err)
		}
		assets[asset] = struct{}{}
		pkgs[env] = pkgWithEnv
	}
	return pkgs, assets, nil
}

func (ctrl *Controller) getChecksum(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, pkg *config.Package, checksumFiles map[string]struct{}, rt *runtime.Runtime, assetNames map[string]struct{}) error { //nolint:funlen,cyclop
	pkgInfo := pkg.PackageInfo

	if !pkg.PackageInfo.Checksum.GetEnabled() {
		if err := ctrl.dlAssetAndGetChecksum(ctx, logE, checksums, pkg, rt); err != nil {
			return err
		}
		return nil
	}

	checksumID, err := pkg.GetChecksumID(rt)
	if err != nil {
		return fmt.Errorf("get a checksum id: %w", err)
	}

	if a := checksums.Get(checksumID); a != nil {
		return nil
	}

	checksumFileID, err := pkg.RenderChecksumFileID(rt)
	if err != nil {
		return fmt.Errorf("render a checksum file ID: %w", err)
	}
	if _, ok := checksumFiles[checksumFileID]; ok {
		return nil
	}
	checksumFiles[checksumFileID] = struct{}{}
	logE.Debug("downloading a checksum file")
	file, _, err := ctrl.chkDL.DownloadChecksum(ctx, logE, rt, pkg)
	if err != nil {
		return fmt.Errorf("download a checksum file: %w", err)
	}
	if file == nil {
		return nil
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
		return nil
	}
	m, s, err := checksum.ParseChecksumFile(checksumFile, pkgInfo.Checksum)
	if err != nil {
		return fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		logE.WithFields(logrus.Fields{
			"checksum_id": checksumID,
			"checksum":    s,
		}).Debug("set a checksum")
		checksums.Set(checksumID, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  s,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
		return nil
	}
	for assetName, chksum := range m {
		if _, ok := assetNames[assetName]; !ok {
			continue
		}
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
	return nil
}

func (ctrl *Controller) dlAssetAndGetChecksum(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, pkg *config.Package, rt *runtime.Runtime) (gErr error) {
	checksumID, err := pkg.GetChecksumID(rt)
	if err != nil {
		return fmt.Errorf("get a checksum id: %w", err)
	}
	if a := checksums.Get(checksumID); a != nil {
		return nil
	}
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return fmt.Errorf("get an asset name: %w", err)
	}
	fields := logrus.Fields{
		"asset_name": assetName,
	}
	logE = logE.WithFields(fields)
	logE.Info("downloading an asset to calculate the checksum")
	defer func() {
		if gErr != nil {
			gErr = logerr.WithFields(gErr, fields)
		}
	}()
	f, err := download.ConvertPackageToFile(pkg, assetName, rt)
	if err != nil {
		return err //nolint:wrapcheck
	}
	file, _, err := ctrl.downloader.GetReadCloser(ctx, logE, f)
	if err != nil {
		return fmt.Errorf("download an asset: %w", err)
	}
	defer file.Close()
	algorithm := "sha512"
	fields["algorithm"] = algorithm
	chk, err := checksum.CalculateReader(file, algorithm)
	if err != nil {
		return fmt.Errorf("calculate an asset: %w", err)
	}
	checksums.Set(checksumID, &checksum.Checksum{
		ID:        checksumID,
		Checksum:  chk,
		Algorithm: algorithm,
	})
	return nil
}
