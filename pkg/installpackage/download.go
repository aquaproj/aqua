package installpackage

import (
	"context"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/codingsince1985/checksum"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *installer) download(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, dest, assetName string, logE *logrus.Entry) error { //nolint:funlen,cyclop
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

	tempDir, err := afero.TempDir(inst.fs, "", "")
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer inst.fs.RemoveAll(tempDir) //nolint:errcheck
	tempFilePath := filepath.Join(tempDir, assetName)
	file, err := inst.fs.Create(tempFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if _, err := io.Copy(file, body); err != nil {
		return err //nolint:wrapcheck
	}
	sha256, err := checksum.SHA256sum(tempFilePath)
	if err != nil {
		return err //nolint:wrapcheck
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

	return unarchive.Unarchive(&unarchive.File{ //nolint:wrapcheck
		Body:     readFile,
		Filename: assetName,
		Type:     pkgInfo.GetFormat(),
	}, dest, logE, inst.fs)
}
