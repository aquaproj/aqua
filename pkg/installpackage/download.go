package installpackage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *Installer) downloadWithRetry(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    param.Package.Package.Name,
		"package_version": param.Package.Package.Version,
		"registry":        param.Package.Package.Registry,
	})
	if param.Package.PackageInfo.Type == "go" {
		return inst.buildGoPkgFiles(ctx, logE, param)
	}
	retryCount := 0
	for {
		logE.Debug("check if the package is already installed")
		finfo, err := inst.fs.Stat(param.Dest)
		if err != nil { //nolint:nestif
			// file doesn't exist
			if err := inst.download(ctx, logE, param); err != nil {
				if strings.Contains(err.Error(), "file already exists") {
					if retryCount >= maxRetryDownload {
						return err
					}
					retryCount++
					logerr.WithError(logE, err).WithFields(logrus.Fields{
						"retry_count": retryCount,
					}).Info("retry installing the package")
					continue
				}
				return err
			}
			return nil
		}
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", param.Dest)
		}
		return nil
	}
}

func (inst *Installer) download(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error { //nolint:funlen,cyclop
	ppkg := param.Package
	pkg := ppkg.Package
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	pkgInfo := param.Package.PackageInfo

	if pkgInfo.Type == "go_install" {
		return inst.downloadGoInstall(ctx, ppkg, param.Dest, logE)
	}

	logE.Info("download and unarchive the package")

	checksumID, err := ppkg.GetChecksumID(inst.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	var chksum *checksum.Checksum
	if param.Checksums != nil {
		chksum = param.Checksums.Get(checksumID)
		if chksum == nil && !pkgInfo.Checksum.GetEnabled() && param.RequireChecksum {
			return logerr.WithFields(errChecksumIsRequired, logrus.Fields{ //nolint:wrapcheck
				"doc": "https://aquaproj.github.io/docs/reference/codes/001",
			})
		}
	}

	body, cl, err := inst.packageDownloader.GetReadCloser(ctx, ppkg, param.Asset, logE)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	var readBody io.Reader = body

	if param.Checksums != nil {
		readFile, err := inst.verifyChecksum(ctx, logE, &ParamVerifyChecksum{
			ChecksumID: checksumID,
			Checksum:   chksum,
			Checksums:  param.Checksums,
			Pkg:        ppkg,
			AssetName:  param.Asset,
			Body:       body,
		})
		if err != nil {
			return err
		}
		if readFile != nil {
			defer readFile.Close()
		}
		readBody = readFile
	}

	var pOpts *unarchive.ProgressBarOpts
	if inst.progressBar {
		pOpts = &unarchive.ProgressBarOpts{
			ContentLength: cl,
			Description:   fmt.Sprintf("Downloading %s %s", pkg.Name, pkg.Version),
		}
	}

	return unarchive.Unarchive(&unarchive.File{ //nolint:wrapcheck
		Body:     readBody,
		Filename: param.Asset,
		Type:     pkgInfo.GetFormat(),
	}, param.Dest, logE, inst.fs, pOpts)
}

func (inst *Installer) downloadGoInstall(ctx context.Context, pkg *config.Package, dest string, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	goPkgPath := pkgInfo.GetPath() + "@" + pkg.Package.Version
	logE.WithFields(logrus.Fields{
		"gobin":           dest,
		"go_package_path": goPkgPath,
	}).Info("Installing a Go tool")
	if _, err := inst.executor.GoInstall(ctx, goPkgPath, dest); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}

func (inst *Installer) buildGoPkgFiles(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error {
	for _, file := range param.Package.PackageInfo.Files {
		file := file
		if err := inst.buildGoPkgFile(ctx, logE, param, file); err != nil {
			return err
		}
	}
	return nil
}

func (inst *Installer) buildGoPkgFile(ctx context.Context, logE *logrus.Entry, param *DownloadParam, file *registry.File) error {
	pkg := param.Package
	pkgInfo := pkg.PackageInfo
	exePath := inst.getGoPkgExePath(pkg, file)
	dir, err := pkg.RenderDir(file, inst.runtime)
	if err != nil {
		return fmt.Errorf("render file dir: %w", err)
	}
	exeDir := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "src", dir)
	if _, err := inst.fs.Stat(exePath); err == nil {
		return nil
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
	if _, err := inst.executor.GoBuild(ctx, exePath, src, exeDir); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}
