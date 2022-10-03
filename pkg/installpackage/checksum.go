package installpackage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *Installer) extractChecksum(pkg *config.Package, assetName string, checksumFile []byte) (string, error) {
	pkgInfo := pkg.PackageInfo

	if pkgInfo.Checksum.FileFormat == "raw" {
		return strings.TrimSpace(string(checksumFile)), nil
	}

	m, err := inst.checksumFileParser.ParseChecksumFile(string(checksumFile), pkg)
	if err != nil {
		return "", fmt.Errorf("parse a checksum file: %w", err)
	}

	for fileName, chksum := range m {
		if fileName != assetName {
			continue
		}
		return chksum, nil
	}

	return "", nil
}

func (inst *Installer) dlAndExtractChecksum(ctx context.Context, logE *logrus.Entry, pkg *config.Package, assetName string) (string, error) {
	file, _, err := inst.checksumDownloader.DownloadChecksum(ctx, logE, inst.runtime, pkg)
	if err != nil {
		return "", fmt.Errorf("download a checksum file: %w", err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read a checksum file: %w", err)
	}

	return inst.extractChecksum(pkg, assetName, b)
}

type ParamVerifyChecksum struct {
	ChecksumID string
	Checksum   *checksum.Checksum
	Checksums  *checksum.Checksums
	Pkg        *config.Package
	AssetName  string
	Body       io.Reader
}

func (inst *Installer) verifyChecksum(ctx context.Context, logE *logrus.Entry, param *ParamVerifyChecksum) (io.ReadCloser, error) { //nolint:cyclop,funlen
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	checksums := param.Checksums
	chksum := param.Checksum
	checksumID := param.ChecksumID

	tempDir, err := afero.TempDir(inst.fs, "", "")
	if err != nil {
		return nil, fmt.Errorf("create a temporal directory: %w", err)
	}
	defer inst.fs.RemoveAll(tempDir) //nolint:errcheck

	assetName := filepath.Base(param.AssetName)
	tempFilePath := filepath.Join(tempDir, assetName)
	if assetName == "" && (pkgInfo.Type == "github_archive" || pkgInfo.Type == "go") {
		tempFilePath = filepath.Join(tempDir, "archive.tar.gz")
	}
	file, err := inst.fs.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("create a temporal file: %w", logerr.WithFields(err, logrus.Fields{
			"temp_file": tempFilePath,
		}))
	}
	defer file.Close()

	if _, err := io.Copy(file, param.Body); err != nil {
		return nil, err //nolint:wrapcheck
	}

	if chksum == nil && pkgInfo.Checksum.GetEnabled() {
		logE.Info("downloading a checksum file")
		c, err := inst.dlAndExtractChecksum(ctx, logE, pkg, assetName)
		if err != nil {
			return nil, err
		}
		chksum = &checksum.Checksum{
			ID:        checksumID,
			Checksum:  c,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		}
		checksums.Set(checksumID, chksum)
	}

	if chksum != nil {
		chksum.Checksum = strings.ToUpper(chksum.Checksum)
	}

	algorithm := "sha512"
	if chksum != nil {
		algorithm = chksum.Algorithm
	}
	calculatedSum, err := checksum.Calculate(inst.fs, tempFilePath, algorithm)
	if err != nil {
		return nil, fmt.Errorf("calculate a checksum of downloaded file: %w", logerr.WithFields(err, logrus.Fields{
			"temp_file": tempFilePath,
		}))
	}
	calculatedSum = strings.ToUpper(calculatedSum)

	if chksum != nil && calculatedSum != chksum.Checksum {
		return nil, logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
			"actual_checksum":   calculatedSum,
			"expected_checksum": chksum.Checksum,
		})
	}

	if chksum == nil {
		logE.WithFields(logrus.Fields{
			"checksum_id": checksumID,
			"checksum":    calculatedSum,
		}).Debug("set a calculated checksum")
		chksum = &checksum.Checksum{
			ID:        checksumID,
			Checksum:  calculatedSum,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		}
	}
	checksums.Set(checksumID, chksum)

	readFile, err := inst.fs.Open(tempFilePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return readFile, nil
}
