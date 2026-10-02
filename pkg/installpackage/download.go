package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/schollz/progressbar/v3"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) downloadWithRetry(ctx context.Context, logger *slog.Logger, param *DownloadParam) error {
	retryCount := 0
	for {
		logger.Debug("check if the package is already installed")
		finfo, err := os.Stat(param.Dest)
		if err != nil { //nolint:nestif
			// file doesn't exist
			if err := is.download(ctx, logger, param); err != nil {
				if strings.Contains(err.Error(), "file already exists") {
					if retryCount >= maxRetryDownload {
						return err
					}
					retryCount++
					slogerr.WithError(logger, err).Info("retry installing the package",
						"retry_count", retryCount)
					continue
				}
				return err
			}
			pkgPath, err := param.Package.PkgPath(is.runtime)
			if err != nil {
				return fmt.Errorf("get a package path: %w", err)
			}
			if err := is.vacuum.Update(pkgPath, time.Now()); err != nil {
				slogerr.WithError(logger, err).Warn("update the last used datetime")
			}
			return nil
		}
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", param.Dest)
		}
		return nil
	}
}

func (is *Installer) download(ctx context.Context, logger *slog.Logger, param *DownloadParam) error { //nolint:cyclop
	ppkg := param.Package
	pkg := ppkg.Package
	pkgInfo := param.Package.PackageInfo

	if pkgInfo.Type == "go_install" {
		return is.downloadGoInstall(ctx, logger, ppkg, param.Dest)
	}

	if pkgInfo.Type == "cargo" {
		return is.downloadCargo(ctx, logger, ppkg, param.Dest)
	}

	logger.Info("download and unarchive the package")

	file, err := download.ConvertPackageToFile(ppkg, param.Asset, is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	body, cl, err := is.downloader.ReadCloser(ctx, logger, file)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	var pb *progressbar.ProgressBar
	if is.progressBar && cl != 0 {
		pb = progressbar.DefaultBytes(
			cl,
			fmt.Sprintf("Downloading %s %s", pkg.Name, pkg.Version),
		)
	}
	bodyFile := download.NewDownloadedFile(body, pb)
	defer func() {
		if err := bodyFile.Remove(); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary file")
		}
	}()

	if err := is.runVerifiers(ctx, logger, is.assetVerifiers(logger, ppkg, param.Asset, is.runtime, param.Locked), bodyFile.Path); err != nil {
		return err
	}

	if err := is.verifyChecksumWrap(ctx, logger, param, bodyFile); err != nil {
		return err
	}

	return is.unarchive(ctx, logger, param, bodyFile, pkgInfo.GetFormat())
}

// unarchive extracts the asset into a temporary directory and moves it to
// param.Dest only once the extraction has finished.
//
// aqua treats the mere existence of the destination directory as proof that the
// package is installed (see downloadWithRetry), so extracting straight into it
// would publish the directory before it is populated: a concurrent install of
// the same package -- another goroutine of this process, or another aqua process
// -- would skip the download and then fail to find the executable that has not
// been written yet. Renaming a fully extracted directory into place keeps that
// check honest, because the destination only ever appears complete.
func (is *Installer) unarchive(ctx context.Context, logger *slog.Logger, param *DownloadParam, bodyFile *download.DownloadedFile, format string) error {
	// The temporary directory must live under rootDir so that it is on the same
	// filesystem as the destination; renaming across drives fails on Windows.
	tempDir := filepath.Join(is.rootDir, "temp")
	if err := osfile.MkdirAll(tempDir); err != nil {
		return fmt.Errorf("create a temporary directory: %w", err)
	}
	// The directory is renamed into place as is, so it must be created with the
	// permissions a package directory needs. os.MkdirTemp would create it with
	// 0700 and leave the package unreadable to every other user.
	tempDir, err := osfile.MkdirTemp(tempDir)
	if err != nil {
		return fmt.Errorf("create a temporary directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary directory")
		}
	}()

	if err := is.unarchiver.Unarchive(ctx, logger, &unarchive.File{
		Body:     bodyFile,
		Filename: param.Asset,
		Type:     format,
	}, tempDir); err != nil {
		return err //nolint:wrapcheck
	}

	if err := osfile.MkdirAll(filepath.Dir(param.Dest)); err != nil {
		return fmt.Errorf("create the parent directory of the package: %w", err)
	}
	if err := os.Rename(tempDir, param.Dest); err != nil {
		// The rename fails if something else already installed the package. Ask the
		// destination rather than the error, whose text differs per platform.
		if _, e := os.Stat(param.Dest); e == nil {
			logger.Debug("the package has been installed by another process")
			return nil
		}
		return fmt.Errorf("move the unarchived package to the destination: %w", err)
	}
	return nil
}

// assetVerifiers returns the signature verifiers to run on the downloaded asset.
//
// A package installed from the lock file gets none of them. Its signatures were
// verified when the entry was written, and the checksum recorded then pins the
// download to the very artifact they signed, so checking them again proves nothing
// the checksum hasn't already. It costs a great deal: cosign, slsa-verifier and gh
// are themselves downloaded and run, against services that are reachable often
// rather than always, which is where installs fail for reasons that have nothing to
// do with the package.
//
// Someone who would rather not take that on trust asks for it back with
// AQUA_VERIFY_SIGNATURES, and then an install verifies exactly as it would without a
// lock file.
func (is *Installer) assetVerifiers(logger *slog.Logger, pkg *config.Package, assetName string, rt *runtime.Runtime, locked bool) []FileVerifier {
	if locked && !is.verifySignatures {
		logger.Debug("skip verifying signatures because the package is installed from the lock file")
		return nil
	}
	pkgInfo := pkg.PackageInfo
	return []FileVerifier{
		&gitHubArtifactAttestationsVerifier{
			disabled:    is.gaaDisabled,
			gaa:         pkgInfo.GitHubArtifactAttestations,
			pkg:         pkg,
			ghInstaller: is.ghInstaller,
			ghVerifier:  is.ghVerifier,
		},
		&cosignVerifier{
			disabled:  is.cosignDisabled,
			pkg:       pkg,
			cosign:    pkgInfo.Cosign,
			installer: is.cosignInstaller,
			verifier:  is.cosign,
			runtime:   rt,
			asset:     assetName,
		},
		&slsaVerifier{
			disabled:  is.slsaDisabled,
			pkg:       pkg,
			installer: is.slsaVerifierInstaller,
			verifier:  is.slsaVerifier,
			runtime:   rt,
			asset:     assetName,
		},
		&minisignVerifier{
			pkg:         pkg,
			installer:   is.minisignInstaller,
			verifier:    is.minisignVerifier,
			runtime:     rt,
			realRuntime: is.realRuntime,
			asset:       assetName,
			minisign:    pkgInfo.Minisign,
		},
	}
}

// runVerifiers runs the verifiers that are enabled against the asset.
//
// path produces the file to verify, and is called only once a verifier turns out to
// be enabled: the download is held in memory until something needs it on disk, and a
// package with no signatures should never pay for a temporary file.
func (is *Installer) runVerifiers(ctx context.Context, logger *slog.Logger, verifiers []FileVerifier, path func() (string, error)) error {
	filePath := ""
	for _, verifier := range verifiers {
		a, err := verifier.Enabled(logger)
		if err != nil {
			return fmt.Errorf("check if the verifier is enabled: %w", err)
		}
		if !a {
			continue
		}
		if filePath == "" {
			a, err := path()
			if err != nil {
				return fmt.Errorf("get the path of the asset to verify: %w", err)
			}
			filePath = a
		}
		if err := verifier.Verify(ctx, logger, filePath); err != nil {
			return fmt.Errorf("verify the asset: %w", err)
		}
	}
	return nil
}

// VerifyAsset verifies the signatures over an asset that has already been downloaded.
//
// It is how the lock file comes to be worth trusting: the checksum recorded for an
// entry is the checksum of an artifact whose signatures were checked here, once,
// which is what lets an install verify the checksum alone.
func (is *Installer) VerifyAsset(ctx context.Context, logger *slog.Logger, pkg *config.Package, assetName, filePath string, rt *runtime.Runtime) error {
	return is.runVerifiers(ctx, logger, is.assetVerifiers(logger, pkg, assetName, rt, false), func() (string, error) {
		return filePath, nil
	})
}
