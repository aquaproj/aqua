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
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// hasChecksumSignatureVerification returns true if the checksum has signature verification configured
// (Cosign, Minisign, or GitHubArtifactAttestations).
func hasChecksumSignatureVerification(chksum *registry.Checksum) bool {
	if chksum == nil {
		return false
	}
	return chksum.GetCosign() != nil || chksum.GetMinisign() != nil || chksum.GetGitHubArtifactAttestations() != nil
}

func (is *Installer) dlAndExtractChecksum(ctx context.Context, logger *slog.Logger, pkg *config.Package, assetName string) (string, error) { //nolint:funlen
	file, _, err := is.checksumDownloader.DownloadChecksum(ctx, logger, is.runtime, pkg)
	if err != nil {
		return "", fmt.Errorf("download a checksum file: %w", err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read a checksum file: %w", err)
	}

	var tempFilePath string
	pkgInfo := pkg.PackageInfo

	verifiers := []FileVerifier{
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

	for _, verifier := range verifiers {
		a, err := verifier.Enabled(logger)
		if err != nil {
			return "", fmt.Errorf("check if the verifier is enabled: %w", err)
		}
		if !a {
			continue
		}
		if tempFilePath == "" {
			f, err := afero.TempFile(is.fs, "", "")
			if err != nil {
				return "", fmt.Errorf("create a temporary file: %w", err)
			}
			tempFilePath = f.Name()
			defer f.Close()
			defer is.fs.Remove(tempFilePath) //nolint:errcheck
			if _, err := f.Write(b); err != nil {
				return "", fmt.Errorf("write a checksum to a temporary file: %w", err)
			}
		}
		if err := verifier.Verify(ctx, logger, tempFilePath); err != nil {
			return "", fmt.Errorf("verify the checksum file: %w", err)
		}
	}

	return checksum.GetChecksum(logger, assetName, string(b), pkg.PackageInfo.Checksum) //nolint:wrapcheck
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

// getChecksumFromSource retrieves the checksum from GitHub API or checksum file.
// It tries GitHub API first if no signature verification is configured, then falls back to checksum file.
func (is *Installer) getChecksumFromSource(ctx context.Context, logger *slog.Logger, pkg *config.Package, assetName, checksumID string) (*checksum.Checksum, error) {
	pkgInfo := pkg.PackageInfo
	hasSignatureVerification := hasChecksumSignatureVerification(pkgInfo.Checksum)

	// If no signature verification and it's a github_release type, try GitHub API first
	if !hasSignatureVerification && pkgInfo.Type == config.PkgInfoTypeGitHubRelease {
		releaseAssets, err := is.checksumDownloader.GetReleaseAssets(ctx, logger, pkg)
		if err != nil {
			slogerr.WithError(logger, err).Debug("failed to get release assets from GitHub API")
		} else if releaseAssets != nil {
			if digest := releaseAssets.GetDigest(assetName); digest != nil {
				logger.Debug("got digest from GitHub API",
					"checksum_id", checksumID,
					"checksum", digest.Digest)
				return &checksum.Checksum{
					ID:        checksumID,
					Checksum:  digest.Digest,
					Algorithm: digest.Algorithm,
				}, nil
			}
		}
	}

	// Fall back to checksum file download
	logger.Info("downloading a checksum file")
	// For github_content, use the base name
	dlAssetName := assetName
	if dlAssetName != "" {
		dlAssetName = filepath.Base(dlAssetName)
	}
	c, err := is.dlAndExtractChecksum(ctx, logger, pkg, dlAssetName)
	if err != nil {
		return nil, slogerr.With(err, "asset_name", dlAssetName) //nolint:wrapcheck
	}
	return &checksum.Checksum{
		ID:        checksumID,
		Checksum:  c,
		Algorithm: pkgInfo.Checksum.GetAlgorithm(),
	}, nil
}

func (is *Installer) verifyChecksum(ctx context.Context, logger *slog.Logger, param *ParamVerifyChecksum) error {
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	checksums := param.Checksums
	chksum := param.Checksum
	checksumID := param.ChecksumID
	tempFilePath := param.TempFilePath

	if chksum == nil && pkgInfo.Checksum.GetEnabled() {
		var err error
		chksum, err = is.getChecksumFromSource(ctx, logger, pkg, param.AssetName, checksumID)
		if err != nil {
			return err
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
