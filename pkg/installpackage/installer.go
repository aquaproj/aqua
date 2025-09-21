package installpackage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"golang.org/x/sync/errgroup"
)

const (
	proxyName        = "aqua-proxy"
	maxRetryDownload = 1
)

type Installer struct {
	downloader            download.ClientAPI
	checksumDownloader    download.ChecksumDownloader
	checksumCalculator    ChecksumCalculator
	linker                Linker
	unarchiver            Unarchiver
	cosign                CosignVerifier
	slsaVerifier          SLSAVerifier
	minisignVerifier      MinisignVerifier
	ghVerifier            GitHubArtifactAttestationsVerifier
	cosignInstaller       *DedicatedInstaller
	slsaVerifierInstaller *DedicatedInstaller
	minisignInstaller     *DedicatedInstaller
	ghInstaller           *DedicatedInstaller
	goInstallInstaller    GoInstallInstaller
	goBuildInstaller      GoBuildInstaller
	cargoPackageInstaller CargoPackageInstaller
	runtime               *runtime.Runtime
	realRuntime           *runtime.Runtime
	fs                    afero.Fs
	rootDir               string
	copyDir               string
	maxParallelism        int
	progressBar           bool
	onlyLink              bool
	cosignDisabled        bool
	slsaDisabled          bool
	gaaDisabled           bool
	graDisabled           bool
	vacuum                Vacuum
}

type Vacuum interface {
	Update(pkgPath string, timestamp time.Time) error
}

func New(param *config.Param, downloader download.ClientAPI, rt *runtime.Runtime, fs afero.Fs, linker Linker, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, cosignVerifier CosignVerifier, slsaVerifier SLSAVerifier, minisignVerifier MinisignVerifier, ghVerifier GitHubArtifactAttestationsVerifier, goInstallInstaller GoInstallInstaller, goBuildInstaller GoBuildInstaller, cargoPackageInstaller CargoPackageInstaller, vacuum Vacuum) *Installer {
	ni := func(rt *runtime.Runtime) *Installer {
		return newInstaller(param, downloader, rt, fs, linker, chkDL, chkCalc, unarchiver, cosignVerifier, slsaVerifier, minisignVerifier, ghVerifier, goInstallInstaller, goBuildInstaller, cargoPackageInstaller, vacuum)
	}
	installer := ni(rt)
	installer.cosignInstaller = newDedicatedInstaller(
		ni(runtime.NewR()),
		cosign.Package,
		cosign.Checksums(),
	)
	installer.slsaVerifierInstaller = newDedicatedInstaller(
		ni(runtime.NewR()),
		slsa.Package,
		slsa.Checksums(),
	)
	installer.minisignInstaller = newDedicatedInstaller(
		ni(runtime.NewR()),
		minisign.Package,
		minisign.Checksums(),
	)
	installer.ghInstaller = newDedicatedInstaller(
		ni(runtime.NewR()),
		ghattestation.Package,
		ghattestation.Checksums(),
	)
	return installer
}

func newInstaller(param *config.Param, downloader download.ClientAPI, rt *runtime.Runtime, fs afero.Fs, linker Linker, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, cosignVerifier CosignVerifier, slsaVerifier SLSAVerifier, minisignVerifier MinisignVerifier, ghVerifier GitHubArtifactAttestationsVerifier, goInstallInstaller GoInstallInstaller, goBuildInstaller GoBuildInstaller, cargoPackageInstaller CargoPackageInstaller, vacuum Vacuum) *Installer {
	return &Installer{
		rootDir:               param.RootDir,
		maxParallelism:        param.MaxParallelism,
		downloader:            downloader,
		checksumDownloader:    chkDL,
		checksumCalculator:    chkCalc,
		runtime:               rt,
		realRuntime:           runtime.NewR(),
		fs:                    fs,
		linker:                linker,
		progressBar:           param.ProgressBar,
		onlyLink:              param.OnlyLink,
		cosignDisabled:        param.CosignDisabled,
		slsaDisabled:          param.SLSADisabled,
		gaaDisabled:           param.GitHubArtifactAttestationDisabled,
		graDisabled:           param.GitHubReleaseAttestationDisabled,
		copyDir:               param.Dest,
		unarchiver:            unarchiver,
		cosign:                cosignVerifier,
		slsaVerifier:          slsaVerifier,
		minisignVerifier:      minisignVerifier,
		ghVerifier:            ghVerifier,
		goInstallInstaller:    goInstallInstaller,
		goBuildInstaller:      goBuildInstaller,
		cargoPackageInstaller: cargoPackageInstaller,
		vacuum:                vacuum,
	}
}

type Linker interface {
	Lstat(s string) (os.FileInfo, error)
	Symlink(dest, src string) error
	Hardlink(dest, src string) error
	Readlink(src string) (string, error)
}

type SLSAVerifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *slsa.ParamVerify) error
}

type MinisignVerifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, m *registry.Minisign, art *template.Artifact, file *download.File, param *minisign.ParamVerify) error
}

type GitHubArtifactAttestationsVerifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ghattestation.ParamVerify) error
	VerifyRelease(ctx context.Context, logE *logrus.Entry, param *ghattestation.ParamVerifyRelease) error
}

type CosignVerifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error
}

type Unarchiver interface {
	Unarchive(ctx context.Context, logE *logrus.Entry, src *unarchive.File, dest string) error
}

type ParamInstallPackages struct {
	ConfigFilePath  string
	Config          *aqua.Config
	Registries      map[string]*registry.Config
	Tags            map[string]struct{}
	ExcludedTags    map[string]struct{}
	PolicyConfigs   []*policy.Config
	Checksums       *checksum.Checksums
	SkipLink        bool
	RequireChecksum bool
	DisablePolicy   bool
}

type ParamInstallPackage struct {
	Pkg             *config.Package
	Checksums       *checksum.Checksums
	RequireChecksum bool
	PolicyConfigs   []*policy.Config
	DisablePolicy   bool
	ConfigFileDir   string
	CosignExePath   string
	Checksum        *checksum.Checksum
}

type ChecksumCalculator interface {
	Calculate(fs afero.Fs, filename, algorithm string) (string, error)
}

func (is *Installer) SetCopyDir(copyDir string) {
	is.copyDir = copyDir
}

type DownloadParam struct {
	Package         *config.Package
	Checksums       *checksum.Checksums
	Checksum        *checksum.Checksum
	Dest            string
	Asset           string
	RequireChecksum bool
}

func (is *Installer) InstallPackages(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackages) error { //nolint:cyclop
	pkgs, failed := config.ListPackages(logE, param.Config, is.runtime, param.Registries)
	if !param.SkipLink {
		if failedCreateLinks := is.createLinks(logE, pkgs); failedCreateLinks {
			failed = failedCreateLinks
		}
	}

	if is.onlyLink {
		logE.WithFields(logrus.Fields{
			"only_link": true,
		}).Debug("skip downloading the package")
		if failed {
			return errInstallFailure
		}
		return nil
	}

	if len(pkgs) == 0 {
		if failed {
			return errInstallFailure
		}
		return nil
	}

	eg := &errgroup.Group{}
	eg.SetLimit(is.maxParallelism)

	for _, pkg := range pkgs {
		logE := logE.WithFields(logrus.Fields{
			"package_name":    pkg.Package.Name,
			"package_version": pkg.Package.Version,
			"registry":        pkg.Package.Registry,
		})
		if !aqua.FilterPackageByTag(pkg.Package, param.Tags, param.ExcludedTags) {
			logE.Debug("skip installing the package because package tags are unmatched")
			continue
		}
		eg.Go(func() error {
			if err := is.InstallPackage(ctx, logE, &ParamInstallPackage{
				Pkg:             pkg,
				Checksums:       param.Checksums,
				RequireChecksum: param.RequireChecksum,
				PolicyConfigs:   param.PolicyConfigs,
				DisablePolicy:   param.DisablePolicy,
			}); err != nil {
				logerr.WithError(logE, err).Error("install the package")
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return errInstallFailure
	}
	return nil
}

func (is *Installer) InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error {
	pkg := param.Pkg
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	logE.Debug("installing the package")

	if err := is.validatePackage(logE, param); err != nil {
		return err
	}

	assetName, err := pkg.RenderAsset(is.runtime)
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	pkgPath, err := pkg.AbsPkgPath(is.rootDir, is.runtime)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}

	if err := is.downloadWithRetry(ctx, logE, &DownloadParam{
		Package:         pkg,
		Dest:            pkgPath,
		Asset:           assetName,
		Checksums:       param.Checksums,
		RequireChecksum: param.RequireChecksum,
		Checksum:        param.Checksum,
	}); err != nil {
		return err
	}

	return is.checkFilesWrap(ctx, logE, param, pkgPath)
}
