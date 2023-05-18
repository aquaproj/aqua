package installpackage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *InstallerImpl) dlAndExtractChecksum(ctx context.Context, logE *logrus.Entry, pkg *config.Package, assetName string) (string, error) {
	file, _, err := inst.checksumDownloader.DownloadChecksum(ctx, logE, inst.runtime, pkg)
	if err != nil {
		return "", fmt.Errorf("download a checksum file: %w", err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read a checksum file: %w", err)
	}

	if cos := pkg.PackageInfo.Checksum.GetCosign(); cos.GetEnabled() {
		f, err := afero.TempFile(inst.fs, "", "")
		if err != nil {
			return "", fmt.Errorf("create a temporal file: %w", err)
		}
		defer f.Close()
		defer inst.fs.Remove(f.Name()) //nolint:errcheck
		if _, err := f.Write(b); err != nil {
			return "", fmt.Errorf("write a checksum to a temporal file: %w", err)
		}
		art := pkg.GetTemplateArtifact(inst.runtime, assetName)
		logE.Info("verify a checksum file with Cosign")
		if err := inst.cosignInstaller.installCosign(ctx, logE, cosign.Version); err != nil {
			return "", err
		}
		if err := inst.cosign.Verify(ctx, logE, inst.runtime, &download.File{
			RepoOwner: pkg.PackageInfo.RepoOwner,
			RepoName:  pkg.PackageInfo.RepoName,
			Version:   pkg.Package.Version,
		}, cos, art, f.Name()); err != nil {
			return "", fmt.Errorf("verify a checksum file with Cosign: %w", err)
		}
	}

	return checksum.GetChecksum(logE, assetName, string(b), pkg.PackageInfo.Checksum) //nolint:wrapcheck
}

type ParamVerifyChecksum struct {
	ChecksumID      string
	Checksum        *checksum.Checksum
	Checksums       *checksum.Checksums
	Pkg             *config.Package
	AssetName       string
	TempFilePath    string
	SkipSetChecksum bool
}

func (inst *InstallerImpl) verifyChecksum(ctx context.Context, logE *logrus.Entry, param *ParamVerifyChecksum) error { //nolint:cyclop,funlen
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	checksums := param.Checksums
	chksum := param.Checksum
	checksumID := param.ChecksumID
	tempFilePath := param.TempFilePath

	// Download an asset in a temporal directory
	// Calculate the checksum of download asset
	// Download a checksum file
	// Extract the checksum from the checksum file
	// Compare the checksum
	// Store the checksum to aqua-checksums.json

	assetName := param.AssetName
	// If pkgInfo.Type is "github_archive", AssetName is empty.
	// filepath.Base("") returns "."
	if assetName != "" {
		// For github_content
		assetName = filepath.Base(assetName)
	}

	if chksum == nil && pkgInfo.Checksum.GetEnabled() {
		logE.Info("downloading a checksum file")
		c, err := inst.dlAndExtractChecksum(ctx, logE, pkg, assetName)
		if err != nil {
			return logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
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

	algorithm := "sha512"
	if chksum != nil {
		algorithm = chksum.Algorithm
	}
	calculatedSum, err := inst.checksumCalculator.Calculate(inst.fs, tempFilePath, algorithm)
	if err != nil {
		return fmt.Errorf("calculate a checksum of downloaded file: %w", logerr.WithFields(err, logrus.Fields{
			"temp_file": tempFilePath,
		}))
	}

	if chksum != nil && !strings.EqualFold(calculatedSum, chksum.Checksum) {
		return logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
			"actual_checksum":   strings.ToUpper(calculatedSum),
			"expected_checksum": strings.ToUpper(chksum.Checksum),
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
	if !param.SkipSetChecksum {
		checksums.Set(checksumID, chksum)
	}
	return nil
}
