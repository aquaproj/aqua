package installpackage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) newChecksumVerifiers(pkg *config.Package, assetName string) []FileVerifier {
	pkgInfo := pkg.PackageInfo
	return []FileVerifier{
		&gitHubArtifactAttestationsVerifier{
			disabled:    is.gaaDisabled,
			gaa:         pkgInfo.Checksum.GetGitHubArtifactAttestations(),
			pkg:         pkg,
			ghInstaller: is.ghInstaller,
			ghVerifier:  is.ghVerifier,
		},
		&gitHubReleaseAttestationsVerifier{
			disabled:    is.graDisabled,
			gra:         pkgInfo.GitHubImmutableRelease,
			pkg:         pkg,
			ghInstaller: is.ghInstaller,
			ghVerifier:  is.ghVerifier,
		},
		&cosignVerifier{
			disabled:  is.cosignDisabled,
			cosign:    pkgInfo.Checksum.GetCosign(),
			pkg:       pkg,
			installer: is.cosignInstaller,
			verifier:  is.cosign,
			runtime:   is.runtime,
			asset:     assetName,
		},
		&minisignVerifier{
			pkg:       pkg,
			minisign:  pkgInfo.Checksum.GetMinisign(),
			installer: is.minisignInstaller,
			verifier:  is.minisignVerifier,
			runtime:   is.runtime,
			asset:     assetName,
		},
	}
}

func (is *Installer) filterVerifiers(logger *slog.Logger, verifiers []FileVerifier) ([]FileVerifier, error) {
	arr := make([]FileVerifier, 0, len(verifiers))
	for _, verifier := range verifiers {
		a, err := verifier.Enabled(logger)
		if err != nil {
			return nil, fmt.Errorf("check if the verifier is enabled: %w", err)
		}
		if !a {
			continue
		}
		arr = append(arr, verifier)
	}
	return arr, nil
}

func (is *Installer) dlAndExtractChecksum(ctx context.Context, logger *slog.Logger, pkg *config.Package, assetName string) (string, error) {
	file, _, err := is.checksumDownloader.DownloadChecksum(ctx, logger, is.runtime, pkg)
	if err != nil {
		return "", fmt.Errorf("download a checksum file: %w", err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read a checksum file: %w", err)
	}

	verifiers, err := is.filterVerifiers(logger, is.newChecksumVerifiers(pkg, assetName))
	if err != nil {
		return "", fmt.Errorf("filter verifiers: %w", err)
	}
	if err := is.verifyChecksumFile(ctx, logger, b, verifiers); err != nil {
		return "", fmt.Errorf("verify the checksum file: %w", err)
	}

	return checksum.GetChecksum(logger, assetName, string(b), pkg.PackageInfo.Checksum) //nolint:wrapcheck
}

func (is *Installer) verifyChecksumFile(ctx context.Context, logger *slog.Logger, b []byte, verifiers []FileVerifier) error {
	if len(verifiers) == 0 {
		return nil
	}
	f, err := afero.TempFile(is.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporary file: %w", err)
	}
	tempFilePath := f.Name()
	defer f.Close()
	defer is.fs.Remove(tempFilePath) //nolint:errcheck
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("write a checksum to a temporary file: %w", err)
	}
	for _, verifier := range verifiers {
		if err := verifier.Verify(ctx, logger, tempFilePath); err != nil {
			return fmt.Errorf("verify the checksum file: %w", err)
		}
	}
	return nil
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

func (is *Installer) verifyChecksumWrap(ctx context.Context, logger *slog.Logger, param *DownloadParam, bodyFile *download.DownloadedFile) error {
	if param.Checksum == nil && param.Checksums == nil {
		return nil
	}
	ppkg := param.Package
	tempFilePath, err := bodyFile.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}
	paramVerifyChecksum := &ParamVerifyChecksum{
		Checksum:        param.Checksum,
		Checksums:       param.Checksums,
		Pkg:             ppkg,
		AssetName:       param.Asset,
		TempFilePath:    tempFilePath,
		SkipSetChecksum: true,
	}

	if param.Checksum == nil {
		paramVerifyChecksum.SkipSetChecksum = false
		cid, err := ppkg.ChecksumID(is.runtime)
		if err != nil {
			return err //nolint:wrapcheck
		}
		paramVerifyChecksum.ChecksumID = cid
		// Even if SLSA Provenance is enabled checksum verification is run
		paramVerifyChecksum.Checksum = param.Checksums.Get(cid)
		if paramVerifyChecksum.Checksum == nil && param.RequireChecksum {
			return slogerr.With(errChecksumIsRequired, //nolint:wrapcheck
				"doc", "https://aquaproj.github.io/docs/reference/codes/001")
		}
	}

	if err := is.verifyChecksum(ctx, logger, paramVerifyChecksum); err != nil {
		return err
	}
	return nil
}

func (is *Installer) verifyChecksum(ctx context.Context, logger *slog.Logger, param *ParamVerifyChecksum) error { //nolint:cyclop
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	checksums := param.Checksums
	chksum := param.Checksum
	checksumID := param.ChecksumID
	tempFilePath := param.TempFilePath

	// Download an asset in a temporary directory
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
		logger.Info("downloading a checksum file")
		c, err := is.dlAndExtractChecksum(ctx, logger, pkg, assetName)
		if err != nil {
			return slogerr.With(err, //nolint:wrapcheck
				"asset_name", assetName)
		}
		chksum = &checksum.Checksum{
			ID:        checksumID,
			Checksum:  c,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		}
		checksums.Set(checksumID, chksum)
	}

	algorithm := "sha256"
	if chksum != nil {
		algorithm = chksum.Algorithm
	}
	calculatedSum, err := is.checksumCalculator.Calculate(is.fs, tempFilePath, algorithm)
	if err != nil {
		return fmt.Errorf("calculate a checksum of downloaded file: %w", slogerr.With(err,
			"temp_file", tempFilePath))
	}

	if chksum != nil && !strings.EqualFold(calculatedSum, chksum.Checksum) {
		return slogerr.With(errInvalidChecksum, //nolint:wrapcheck
			"actual_checksum", strings.ToUpper(calculatedSum),
			"expected_checksum", strings.ToUpper(chksum.Checksum))
	}

	if chksum == nil {
		logger.Debug("set a calculated checksum",
			"checksum_id", checksumID,
			"checksum", calculatedSum)
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
