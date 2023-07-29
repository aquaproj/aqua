package installpackage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const proxyName = "aqua-proxy"

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

func (inst *InstallerImpl) SetCopyDir(copyDir string) {
	inst.copyDir = copyDir
}

func (inst *InstallerImpl) InstallPackages(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackages) error { //nolint:funlen,cyclop
	pkgs, failed := config.ListPackages(logE, param.Config, inst.runtime, param.Registries)
	if !param.SkipLink {
		if failedCreateLinks := inst.createLinks(logE, pkgs); failedCreateLinks {
			failed = failedCreateLinks
		}
	}

	if inst.onlyLink {
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
	maxInstallChan := make(chan struct{}, inst.maxParallelism)

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
			if err := inst.InstallPackage(ctx, logE, &ParamInstallPackage{
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

func (inst *InstallerImpl) InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error { //nolint:cyclop,funlen
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	logE.Debug("install the package")

	if pkgInfo.NoAsset != nil && *pkgInfo.NoAsset {
		logE.Error(fmt.Sprintf("failed to install a package %s@%s. No asset is released in this version", pkg.Package.Name, pkg.Package.Version))
		return errors.New("")
	}
	if pkgInfo.ErrorMessage != "" {
		logE.Error(fmt.Sprintf("failed to install a package %s@%s. %s", pkg.Package.Name, pkg.Package.Version, pkgInfo.ErrorMessage))
		return errors.New("")
	}

	if !param.DisablePolicy {
		if err := inst.policyChecker.ValidatePackage(logE, param.Pkg, param.PolicyConfigs); err != nil {
			return err //nolint:wrapcheck
		}
	}

	if err := pkgInfo.Validate(); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	if pkgInfo.Type == "go_install" && pkg.Package.Version == "latest" {
		return errGoInstallForbidLatest
	}

	assetName, err := pkg.RenderAsset(inst.runtime)
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}

	if err := inst.downloadWithRetry(ctx, logE, &DownloadParam{
		Package:         pkg,
		Dest:            pkgPath,
		Asset:           assetName,
		Checksums:       param.Checksums,
		RequireChecksum: param.RequireChecksum,
		Checksum:        param.Checksum,
	}); err != nil {
		return err
	}

	failed := false
	notFound := false
	for _, file := range pkgInfo.GetFiles() {
		logE := logE.WithField("file_name", file.Name)
		var errFileNotFound *config.FileNotFoundError
		if err := inst.checkAndCopyFile(ctx, pkg, file, logE); err != nil {
			if errors.As(err, &errFileNotFound) {
				notFound = true
			}
			failed = true
			logerr.WithError(logE, err).Error("check file_src is correct")
		}
	}
	if notFound { //nolint:nestif
		paths, err := inst.walk(pkgPath)
		if err != nil {
			logerr.WithError(logE, err).Warn("traverse the content of unarchived package")
		} else {
			if len(paths) > 30 { //nolint:gomnd
				logE.Errorf("executable files aren't found\nFiles in the unarchived package (Only 30 files are shown):\n%s\n ", strings.Join(paths[:30], "\n"))
			} else {
				logE.Errorf("executable files aren't found\nFiles in the unarchived package:\n%s\n ", strings.Join(paths, "\n"))
			}
		}
	}
	if failed {
		return errors.New("check file_src is correct")
	}

	return nil
}

func (inst *InstallerImpl) createLinks(logE *logrus.Entry, pkgs []*config.Package) bool {
	failed := false
	for _, pkg := range pkgs {
		pkgInfo := pkg.PackageInfo
		for _, file := range pkgInfo.GetFiles() {
			if isWindows(inst.runtime.GOOS) {
				if err := inst.createProxyWindows(file.Name, logE); err != nil {
					logerr.WithError(logE, err).Error("create the proxy file")
					failed = true
				}
				continue
			}
			if err := inst.createLink(filepath.Join(inst.rootDir, "bin", file.Name), filepath.Join("..", proxyName), logE); err != nil {
				logerr.WithError(logE, err).Error("create the symbolic link")
				failed = true
				continue
			}
		}
	}
	return failed
}

const maxRetryDownload = 1

type DownloadParam struct {
	Package         *config.Package
	Checksums       *checksum.Checksums
	Checksum        *checksum.Checksum
	Dest            string
	Asset           string
	RequireChecksum bool
}

func (inst *InstallerImpl) checkAndCopyFile(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) error {
	exePath, err := inst.checkFileSrc(ctx, pkg, file, logE)
	if err != nil {
		return fmt.Errorf("check file_src is correct: %w", err)
	}
	if inst.copyDir == "" {
		return nil
	}
	logE.Info("copying an executable file")
	if err := inst.Copy(filepath.Join(inst.copyDir, file.Name), exePath); err != nil {
		return err
	}

	return nil
}

func (inst *InstallerImpl) checkFileSrcGo(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) (string, error) {
	pkgInfo := pkg.PackageInfo
	exePath := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "bin", file.Name)
	if isWindows(inst.runtime.GOOS) {
		exePath += ".exe"
	}
	dir, err := pkg.RenderDir(file, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("render file dir: %w", err)
	}
	exeDir := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "src", dir)
	if _, err := inst.fs.Stat(exePath); err == nil {
		return exePath, nil
	}
	src := file.Src
	if src == "" {
		src = "."
	}
	logE.WithFields(logrus.Fields{
		"exe_path":     exePath,
		"go_src":       src,
		"go_build_dir": exeDir,
	}).Info("building Go tool")
	if err := inst.goBuildInstaller.Install(ctx, exePath, src, exeDir); err != nil {
		return "", fmt.Errorf("build Go tool: %w", err)
	}
	return exePath, nil
}

func (inst *InstallerImpl) checkFileSrc(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) (string, error) {
	if pkg.PackageInfo.Type == "go_build" {
		return inst.checkFileSrcGo(ctx, pkg, file, logE)
	}

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get the package install path: %w", err)
	}

	fileSrc, err := pkg.RenameFile(logE, inst.fs, pkgPath, file, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get file_src: %w", err)
	}

	exePath := filepath.Join(pkgPath, fileSrc)
	finfo, err := inst.fs.Stat(exePath)
	if err != nil {
		return "", fmt.Errorf("exe_path isn't found: %w", logerr.WithFields(&config.FileNotFoundError{
			Err: err,
		}, logE.WithField("exe_path", exePath).Data))
	}
	if finfo.IsDir() {
		return "", logerr.WithFields(errExePathIsDirectory, logE.WithField("exe_path", exePath).Data) //nolint:wrapcheck
	}

	logE.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !util.IsOwnerExecutable(mode) {
		logE.Debug("add the permission to execute the command")
		if err := inst.fs.Chmod(exePath, util.AllowOwnerExec(mode)); err != nil {
			return "", logerr.WithFields(errChmod, logE.Data) //nolint:wrapcheck
		}
	}

	return exePath, nil
}

const (
	filePermission os.FileMode = 0o755
)

func (inst *InstallerImpl) Copy(dest, src string) error {
	dst, err := inst.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer dst.Close()
	srcFile, err := inst.fs.Open(src)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer srcFile.Close()
	if _, err := io.Copy(dst, srcFile); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}

	return nil
}
