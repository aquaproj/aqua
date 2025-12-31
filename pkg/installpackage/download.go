package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/schollz/progressbar/v3"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) downloadWithRetry(ctx context.Context, logger *slog.Logger, param *DownloadParam) error {
	logger = logger.With(
		"package_name", param.Package.Package.Name,
		"package_version", param.Package.Package.Version,
		"registry", param.Package.Package.Registry,
	)
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
	logger = logger.With(
		"package_name", pkg.Name,
		"package_version", pkg.Version,
		"registry", pkg.Registry,
	)
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
		&gitHubReleaseAttestationsVerifier{
			disabled:    is.graDisabled,
			gra:         pkgInfo.GitHubImmutableRelease,
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

	return is.unarchiver.Unarchive(ctx, logger, &unarchive.File{ //nolint:wrapcheck
		Body:     bodyFile,
		Filename: param.Asset,
		Type:     pkgInfo.GetFormat(),
	}, param.Dest)
}
