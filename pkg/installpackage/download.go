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

func (inst *installer) download(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, dest, assetName string, checksumEnabled bool, logE *logrus.Entry) error { //nolint:funlen,cyclop
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Info("download and unarchive the package")

	body, err := inst.packageDownloader.GetReadCloser(ctx, pkg, pkgInfo, assetName, logE)
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
		if assetName == "" && pkgInfo.Type == "github_archive" {
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

		checksumID, err := pkgInfo.GetChecksumID(pkg, inst.runtime)
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

	return unarchive.Unarchive(&unarchive.File{ //nolint:wrapcheck
		Body:     readBody,
		Filename: assetName,
		Type:     pkgInfo.GetFormat(),
	}, dest, logE, inst.fs)
}
