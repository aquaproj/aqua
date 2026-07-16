package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) downloadWithRetry(ctx context.Context, logger *slog.Logger, param *DownloadParam) error {
	retryCount := 0
	for {
		logger.Debug("check if the package is already installed")
		finfo, err := is.fs.Stat(param.Dest)
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

func (is *Installer) download(ctx context.Context, logger *slog.Logger, param *DownloadParam) error { //nolint:funlen,cyclop
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
	bodyFile := download.NewDownloadedFile(is.fs, body, pb)
	defer func() {
		if err := bodyFile.Remove(); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary file")
		}
	}()

	verifiers := []FileVerifier{
		&gitHubArtifactAttestationsVerifier{
			disabled:    is.gaaDisabled,
			gaa:         pkgInfo.GitHubArtifactAttestations,
			pkg:         ppkg,
			ghInstaller: is.ghInstaller,
			ghVerifier:  is.ghVerifier,
		},
		&cosignVerifier{
			disabled:  is.cosignDisabled,
			pkg:       ppkg,
			cosign:    pkgInfo.Cosign,
			installer: is.cosignInstaller,
			verifier:  is.cosign,
			runtime:   is.runtime,
			asset:     param.Asset,
		},
		&slsaVerifier{
			disabled:  is.slsaDisabled,
			pkg:       ppkg,
			installer: is.slsaVerifierInstaller,
			verifier:  is.slsaVerifier,
			runtime:   is.runtime,
			asset:     param.Asset,
		},
		&minisignVerifier{
			pkg:       ppkg,
			installer: is.minisignInstaller,
			verifier:  is.minisignVerifier,
			runtime:   is.runtime,
			asset:     param.Asset,
			minisign:  pkgInfo.Minisign,
		},
	}

	var tempFilePath string
	for _, verifier := range verifiers {
		a, err := verifier.Enabled(logger)
		if err != nil {
			return fmt.Errorf("check if the verifier is enabled: %w", err)
		}
		if !a {
			continue
		}
		if tempFilePath == "" {
			a, err := bodyFile.Path()
			if err != nil {
				return fmt.Errorf("get a temporary file path: %w", err)
			}
			tempFilePath = a
		}
		if err := verifier.Verify(ctx, logger, tempFilePath); err != nil {
			return fmt.Errorf("verify the asset: %w", err)
		}
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
	if err := osfile.MkdirAll(is.fs, tempDir); err != nil {
		return fmt.Errorf("create a temporary directory: %w", err)
	}
	tempDir, err := afero.TempDir(is.fs, tempDir, "")
	if err != nil {
		return fmt.Errorf("create a temporary directory: %w", err)
	}
	defer func() {
		if err := is.fs.RemoveAll(tempDir); err != nil {
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

	if err := osfile.MkdirAll(is.fs, filepath.Dir(param.Dest)); err != nil {
		return fmt.Errorf("create the parent directory of the package: %w", err)
	}
	if err := is.fs.Rename(tempDir, param.Dest); err != nil {
		// The rename fails if something else already installed the package. Ask the
		// destination rather than the error, whose text differs per platform.
		if _, e := is.fs.Stat(param.Dest); e == nil {
			logger.Debug("the package has been installed by another process")
			return nil
		}
		return fmt.Errorf("move the unarchived package to the destination: %w", err)
	}
	return nil
}
