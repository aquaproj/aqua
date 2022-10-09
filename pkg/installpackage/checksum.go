package installpackage

import (
	"context"
	"errors"
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

	m, s, err := inst.checksumFileParser.ParseChecksumFile(string(checksumFile), pkg)
	if err != nil {
		return "", fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		return s, nil
	}

	return m[assetName], nil
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

	c, err := inst.extractChecksum(pkg, assetName, b)
	if err != nil {
		return "", err
	}
	if c == "" {
		return "", errors.New("checksum isn't found in a checksum file")
	}
	return c, nil
}

type ParamVerifyChecksum struct {
	ChecksumID string
	Checksum   *checksum.Checksum
	Checksums  *checksum.Checksums
	Pkg        *config.Package
	AssetName  string
	Body       io.Reader
	TempDir    string
}

func copyAsset(fs afero.Fs, tempFilePath string, body io.Reader) error {
	file, err := fs.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", logerr.WithFields(err, logrus.Fields{
			"temp_file": tempFilePath,
		}))
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func (inst *Installer) verifyChecksum(ctx context.Context, logE *logrus.Entry, param *ParamVerifyChecksum) (io.ReadCloser, error) { //nolint:cyclop,funlen
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	checksums := param.Checksums
	chksum := param.Checksum
	checksumID := param.ChecksumID
	tempDir := param.TempDir

	// Download an asset in a temporal directory
	// Calculate the checksum of download asset
	// Download a checksum file
	// Extract the checksum from the checksum file
	// Compare the checksum
	// Store the checksum to aqua-checksums.json

	assetName := param.AssetName
	tempFilePath := filepath.Join(tempDir, assetName)
	if assetName == "" && (pkgInfo.Type == "github_archive" || pkgInfo.Type == "go") {
		tempFilePath = filepath.Join(tempDir, "archive.tar.gz")
	}
	if err := copyAsset(inst.fs, tempFilePath, param.Body); err != nil {
		return nil, err
	}

	if chksum == nil && pkgInfo.Checksum.GetEnabled() {
		logE.Info("downloading a checksum file")
		c, err := inst.dlAndExtractChecksum(ctx, logE, pkg, assetName)
		if err != nil {
			return nil, logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"asset_name": assetName,
			})
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
	calculatedSum, err := inst.checksumCalculator.Calculate(inst.fs, tempFilePath, algorithm)
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
