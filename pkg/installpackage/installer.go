package installpackage

import (
	"context"
	"fmt"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const (
	proxyName        = "aqua-proxy"
	maxRetryDownload = 1
)

type InstallerImpl struct {
	downloader            download.ClientAPI
	checksumDownloader    download.ChecksumDownloader
	checksumCalculator    ChecksumCalculator
	linker                domain.Linker
	unarchiver            unarchive.Unarchiver
	cosign                cosign.Verifier
	slsaVerifier          slsa.Verifier
	policyChecker         *policy.Checker
	cosignInstaller       *Cosign
	slsaVerifierInstaller *SLSAVerifier
	goInstallInstaller    GoInstallInstaller
	goBuildInstaller      GoBuildInstaller
	cargoPackageInstaller CargoPackageInstaller
	runtime               *runtime.Runtime
	fs                    afero.Fs
	rootDir               string
	copyDir               string
	maxParallelism        int
	progressBar           bool
	onlyLink              bool
}

func New(param *config.Param, downloader download.ClientAPI, rt *runtime.Runtime, fs afero.Fs, linker domain.Linker, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver unarchive.Unarchiver, policyChecker *policy.Checker, cosignVerifier cosign.Verifier, slsaVerifier slsa.Verifier, goInstallInstaller GoInstallInstaller, goBuildInstaller GoBuildInstaller, cargoPackageInstaller CargoPackageInstaller) *InstallerImpl {
	installer := newInstaller(param, downloader, rt, fs, linker, chkDL, chkCalc, unarchiver, policyChecker, cosignVerifier, slsaVerifier, goInstallInstaller, goBuildInstaller, cargoPackageInstaller)
	installer.cosignInstaller = &Cosign{
		installer: newInstaller(param, downloader, runtime.NewR(), fs, linker, chkDL, chkCalc, unarchiver, policyChecker, cosignVerifier, slsaVerifier, goInstallInstaller, goBuildInstaller, cargoPackageInstaller),
		mutex:     &sync.Mutex{},
	}
	installer.slsaVerifierInstaller = &SLSAVerifier{
		installer: newInstaller(param, downloader, runtime.NewR(), fs, linker, chkDL, chkCalc, unarchiver, policyChecker, cosignVerifier, slsaVerifier, goInstallInstaller, goBuildInstaller, cargoPackageInstaller),
		mutex:     &sync.Mutex{},
	}
	return installer
}

func newInstaller(param *config.Param, downloader download.ClientAPI, rt *runtime.Runtime, fs afero.Fs, linker domain.Linker, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver unarchive.Unarchiver, policyChecker *policy.Checker, cosignVerifier cosign.Verifier, slsaVerifier slsa.Verifier, goInstallInstaller GoInstallInstaller, goBuildInstaller GoBuildInstaller, cargoPackageInstaller CargoPackageInstaller) *InstallerImpl {
	return &InstallerImpl{
		rootDir:               param.RootDir,
		maxParallelism:        param.MaxParallelism,
		downloader:            downloader,
		checksumDownloader:    chkDL,
		checksumCalculator:    chkCalc,
		runtime:               rt,
		fs:                    fs,
		linker:                linker,
		progressBar:           param.ProgressBar,
		onlyLink:              param.OnlyLink,
		copyDir:               param.Dest,
		unarchiver:            unarchiver,
		policyChecker:         policyChecker,
		cosign:                cosignVerifier,
		slsaVerifier:          slsaVerifier,
		goInstallInstaller:    goInstallInstaller,
		goBuildInstaller:      goBuildInstaller,
		cargoPackageInstaller: cargoPackageInstaller,
	}
}

type Installer interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackages) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
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

func isWindows(goos string) bool {
	return goos == "windows"
}

func (is *InstallerImpl) SetCopyDir(copyDir string) {
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

func (is *InstallerImpl) InstallPackages(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackages) error { //nolint:funlen,cyclop
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

	var wg sync.WaitGroup
	wg.Add(len(pkgs))
	var flagMutex sync.Mutex
	maxInstallChan := make(chan struct{}, is.maxParallelism)

	handleFailure := func() {
		flagMutex.Lock()
		failed = true
		flagMutex.Unlock()
	}

	for _, pkg := range pkgs {
		go func(pkg *config.Package) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			defer func() {
				<-maxInstallChan
			}()
			logE := logE.WithFields(logrus.Fields{
				"package_name":    pkg.Package.Name,
				"package_version": pkg.Package.Version,
				"registry":        pkg.Package.Registry,
			})
			if !aqua.FilterPackageByTag(pkg.Package, param.Tags, param.ExcludedTags) {
				logE.Debug("skip installing the package because package tags are unmatched")
				return
			}
			if err := is.InstallPackage(ctx, logE, &ParamInstallPackage{
				Pkg:             pkg,
				Checksums:       param.Checksums,
				RequireChecksum: param.Config.RequireChecksum(param.RequireChecksum),
				PolicyConfigs:   param.PolicyConfigs,
			}); err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}
		}(pkg)
	}
	wg.Wait()
	if failed {
		return errInstallFailure
	}
	return nil
}

func (is *InstallerImpl) InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error {
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

	pkgPath, err := pkg.PkgPath(is.rootDir, is.runtime)
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
