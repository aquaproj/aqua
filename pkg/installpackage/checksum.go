package installpackage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/download"
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

	return inst.checksumFileParser.GetChecksum(logE, assetName, string(b), pkg.PackageInfo.Checksum) //nolint:wrapcheck
}

type ParamVerifyChecksum struct {
	ChecksumID      string
	Checksum        *checksum.Checksum
	Checksums       *checksum.Checksums
	Pkg             *config.Package
	AssetName       string
	Body            io.Reader
	TempDir         string
	SkipSetChecksum bool
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

func (inst *InstallerImpl) verifyChecksum(ctx context.Context, logE *logrus.Entry, param *ParamVerifyChecksum) (io.ReadCloser, error) { //nolint:cyclop,funlen
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
	// If pkgInfo.Type is "github_archive", AssetName is empty.
	// filepath.Base("") returns "."
	if assetName != "" {
		// For github_content
		assetName = filepath.Base(assetName)
	}
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

	if chksum != nil && !strings.EqualFold(calculatedSum, chksum.Checksum) {
		return nil, logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
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

	readFile, err := inst.fs.Open(tempFilePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return readFile, nil
}
