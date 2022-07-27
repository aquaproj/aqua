package installpackage

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func (inst *installer) download(ctx context.Context, pkg *config.Package, dest, assetName string, checksumEnabled bool, logE *logrus.Entry) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	pkgInfo := pkg.PackageInfo

	if pkgInfo.Type == "go_install" {
		return inst.downloadGoInstall(ctx, pkg, dest, logE)
	}

	logE.Info("download and unarchive the package")

	body, cl, err := inst.packageDownloader.GetReadCloser(ctx, pkg, assetName, logE)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	var readBody io.Reader = body

	if checksumEnabled {
		readFile, err := inst.verifyChecksum(pkg, assetName, body)
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
			Description:   fmt.Sprintf("Downloading %s %s", pkg.Package.Name, pkg.Package.Version),
		}
	}

	return unarchive.Unarchive(&unarchive.File{ //nolint:wrapcheck
		Body:     readBody,
		Filename: assetName,
		Type:     pkgInfo.GetFormat(),
	}, dest, logE, inst.fs, pOpts)
}

func (inst *installer) downloadGoInstall(ctx context.Context, pkg *config.Package, dest string, logE *logrus.Entry) error {
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
