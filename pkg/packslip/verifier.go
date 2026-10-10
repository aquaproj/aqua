package packslip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type Verifier struct {
	downloader download.ClientAPI
	exe        Executor
}

func New(downloader download.ClientAPI, exe Executor) *Verifier {
	return &Verifier{
		downloader: downloader,
		exe:        exe,
	}
}

type ParamVerify struct {
	BundlePath     string
	ArtifactPath   string
	PublicKey      string
	Identity       string
	IdentityPrefix string
	Issuer         string
	Pin            string
	AllowUnlogged  bool
}

var errNoTrust = errors.New("packslip needs pubkey, identity, identity_prefix, or pin to know whom to trust")

// Verify verifies the downloaded asset with the packslip bundle of the package's release.
//
// assetName is the name the asset is published under. packslip looks an artifact up in
// the signed document by its file name, so the downloaded file is checked under it.
func (v *Verifier) Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, ps *registry.Packslip, art *template.Artifact, file *download.File, artifactPath, assetName string) error {
	if !ps.HasTrust() {
		return errNoTrust
	}
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("create a temporary directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary directory")
		}
	}()

	bundlePath := filepath.Join(tempDir, registry.BundleAssetPackslip)
	if err := v.downloadBundle(ctx, logger, rt, art, file, bundlePath); err != nil {
		return err
	}

	// The bundle is staged in the same directory, but under a name of its own, so an
	// asset can't take its place.
	artifactDir := filepath.Join(tempDir, "artifact")
	if err := os.Mkdir(artifactDir, 0o700); err != nil { //nolint:mnd
		return fmt.Errorf("create a temporary directory: %w", err)
	}
	stagedPath := filepath.Join(artifactDir, filepath.Base(assetName))
	if err := stage(artifactPath, stagedPath); err != nil {
		return err
	}

	return v.exe.Verify(ctx, logger, &ParamVerify{ //nolint:wrapcheck
		BundlePath:     bundlePath,
		ArtifactPath:   stagedPath,
		PublicKey:      ps.PublicKey,
		Identity:       ps.Identity,
		IdentityPrefix: ps.IdentityPrefix,
		Issuer:         ps.Issuer,
		Pin:            ps.Pin,
		AllowUnlogged:  ps.AllowUnlogged,
	})
}

func (v *Verifier) downloadBundle(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, art *template.Artifact, file *download.File, dest string) error {
	asset := registry.BundleAssetPackslip
	f, err := download.ConvertDownloadedFileToFile(&registry.DownloadedFile{
		Type:  file.Type,
		Asset: &asset,
	}, file, rt, art)
	if err != nil {
		return err //nolint:wrapcheck
	}

	rc, _, err := v.downloader.ReadCloser(ctx, logger, f)
	if err != nil {
		return fmt.Errorf("download a packslip bundle: %w", err)
	}
	defer rc.Close()

	bundleFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create a file for a packslip bundle: %w", err)
	}
	defer bundleFile.Close()
	if _, err := io.Copy(bundleFile, rc); err != nil {
		return fmt.Errorf("write a packslip bundle to a file: %w", err)
	}
	return nil
}

// stage puts the downloaded asset at dest, linking it when it can.
func stage(src, dest string) error {
	if err := os.Link(src, dest); err == nil {
		return nil
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open a downloaded asset: %w", err)
	}
	defer srcFile.Close()
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create a file to verify with packslip: %w", err)
	}
	defer destFile.Close()
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("copy a downloaded asset to verify with packslip: %w", err)
	}
	return nil
}
