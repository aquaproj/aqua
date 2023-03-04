package installpackage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/slsa"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const proxyName = "aqua-proxy"

type InstallerImpl struct {
	rootDir            string
	maxParallelism     int
	downloader         download.ClientAPI
	checksumDownloader download.ChecksumDownloader
	checksumFileParser *checksum.FileParser
	checksumCalculator ChecksumCalculator
	runtime            *runtime.Runtime
	fs                 afero.Fs
	linker             domain.Linker
	executor           Executor
	unarchiver         Unarchiver
	cosign             cosign.Verifier
	slsaVerifier       slsa.Verifier
	progressBar        bool
	onlyLink           bool
	copyDir            string
	policyChecker      policy.Checker
	cosignInstaller    *Cosign
}

func New(param *config.Param, downloader download.ClientAPI, rt *runtime.Runtime, fs afero.Fs, linker domain.Linker, executor Executor, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, policyChecker policy.Checker, cosignVerifier cosign.Verifier, slsaVerifier slsa.Verifier) *InstallerImpl {
	installer := newInstaller(param, downloader, rt, fs, linker, executor, chkDL, chkCalc, unarchiver, policyChecker, cosignVerifier, slsaVerifier)
	installer.cosignInstaller = &Cosign{
		installer: newInstaller(param, downloader, runtime.NewR(), fs, linker, executor, chkDL, chkCalc, unarchiver, policyChecker, cosignVerifier, slsaVerifier),
		mutex:     &sync.Mutex{},
	}
	return installer
}

func newInstaller(param *config.Param, downloader download.ClientAPI, rt *runtime.Runtime, fs afero.Fs, linker domain.Linker, executor Executor, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, policyChecker policy.Checker, cosignVerifier cosign.Verifier, slsaVerifier slsa.Verifier) *InstallerImpl {
	return &InstallerImpl{
		rootDir:            param.RootDir,
		maxParallelism:     param.MaxParallelism,
		downloader:         downloader,
		checksumDownloader: chkDL,
		checksumFileParser: &checksum.FileParser{},
		checksumCalculator: chkCalc,
		runtime:            rt,
		fs:                 fs,
		linker:             linker,
		executor:           executor,
		progressBar:        param.ProgressBar,
		onlyLink:           param.OnlyLink,
		copyDir:            param.Dest,
		unarchiver:         unarchiver,
		policyChecker:      policyChecker,
		cosign:             cosignVerifier,
		slsaVerifier:       slsaVerifier,
	}
}

type Installer interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackages) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}

type ParamInstallPackages struct {
	ConfigFilePath string
	Config         *aqua.Config
	Registries     map[string]*registry.Config
	Tags           map[string]struct{}
	ExcludedTags   map[string]struct{}
	SkipLink       bool
	PolicyConfigs  []*policy.Config
	Checksums      *checksum.Checksums
}

type ParamInstallPackage struct {
	Pkg             *config.Package
	Checksums       *checksum.Checksums
	RequireChecksum bool
	PolicyConfigs   []*policy.Config
	ConfigFileDir   string
	CosignExePath   string
	Checksum        *checksum.Checksum
}

type Unarchiver interface {
	Unarchive(src *unarchive.File, dest string, logE *logrus.Entry, fs afero.Fs, prgOpts *unarchive.ProgressBarOpts) error
}

type MockUnarchiver struct {
	Err error
}

func (unarchiver *MockUnarchiver) Unarchive(src *unarchive.File, dest string, logE *logrus.Entry, fs afero.Fs, prgOpts *unarchive.ProgressBarOpts) error {
	return unarchiver.Err
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
		if failedCreateLinks := inst.createLinks(logE, pkgs); !failedCreateLinks {
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
				RequireChecksum: param.Config.RequireChecksum(),
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

func (inst *InstallerImpl) InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error {
	pkg := param.Pkg
	checksums := param.Checksums
	pkgInfo := pkg.PackageInfo
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	logE.Debug("install the package")

	if err := inst.policyChecker.ValidatePackage(&policy.ParamValidatePackage{
		Pkg:           param.Pkg,
		PolicyConfigs: param.PolicyConfigs,
	}); err != nil {
		return err //nolint:wrapcheck
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
		Checksums:       checksums,
		RequireChecksum: param.RequireChecksum,
		Checksum:        param.Checksum,
	}); err != nil {
		return err
	}

	for _, file := range pkgInfo.GetFiles() {
		file := file
		logE := logE.WithField("file_name", file.Name)
		if err := inst.checkAndCopyFile(pkg, file, logE); err != nil {
			return fmt.Errorf("check file_src is correct: %w", err)
		}
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
			if err := inst.createLink(filepath.Join(inst.rootDir, "bin", file.Name), proxyName, logE); err != nil {
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

func (inst *InstallerImpl) checkAndCopyFile(pkg *config.Package, file *registry.File, logE *logrus.Entry) error {
	exePath, err := inst.checkFileSrc(pkg, file, logE)
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

func (inst *InstallerImpl) checkFileSrc(pkg *config.Package, file *registry.File, logE *logrus.Entry) (string, error) {
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
		return "", fmt.Errorf("exe_path isn't found: %w", logerr.WithFields(err, logE.Data))
	}
	if finfo.IsDir() {
		return "", logerr.WithFields(errExePathIsDirectory, logE.Data) //nolint:wrapcheck
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
