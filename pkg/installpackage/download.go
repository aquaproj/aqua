package installpackage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/codingsince1985/checksum"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *installer) download(ctx context.Context, pkg *config.Package, dest, assetName string, checksumEnabled bool, logE *logrus.Entry) error { //nolint:funlen,gocognit,cyclop
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

	if checksumEnabled { //nolint:nestif
		tempDir, err := afero.TempDir(inst.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal directory: %w", err)
		}
		defer inst.fs.RemoveAll(tempDir) //nolint:errcheck
		tempFilePath := filepath.Join(tempDir, assetName)
		if assetName == "" && (pkgInfo.Type == "github_archive" || pkgInfo.Type == "go") {
			tempFilePath = filepath.Join(tempDir, "archive.tar.gz")
		}
		file, err := inst.fs.Create(tempFilePath)
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", logerr.WithFields(err, logrus.Fields{
				"temp_file": tempFilePath,
			}))
		}
		defer file.Close()
		if _, err := io.Copy(file, body); err != nil {
			return err //nolint:wrapcheck
		}
		sha256, err := checksum.SHA256sum(tempFilePath)
		if err != nil {
			return fmt.Errorf("calculate a checksum of downloaded file: %w", logerr.WithFields(err, logrus.Fields{
				"temp_file": tempFilePath,
			}))
		}

		checksumID, err := pkg.GetChecksumID(inst.runtime)
		if err != nil {
			return err //nolint:wrapcheck
		}
		chksum := inst.checksums.Get(checksumID)

		if chksum != "" && sha256 != chksum {
			return logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
				"actual_checksum":   sha256,
				"expected_checksum": chksum,
			})
		}
		if chksum == "" {
			inst.checksums.Set(checksumID, sha256)
		}
		readFile, err := inst.fs.Open(tempFilePath)
		if err != nil {
			return err //nolint:wrapcheck
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
