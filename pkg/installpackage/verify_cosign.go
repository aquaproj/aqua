package installpackage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type cosignVerifier struct {
	disabled  bool
	pkg       *config.Package
	cosign    *registry.Cosign
	installer *DedicatedInstaller
	verifier  CosignVerifier
	runtime   *runtime.Runtime
	asset     string
}

func (*cosignVerifier) Kind() string {
	return "cosign"
}

func (c *cosignVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if c.disabled {
		logger.Debug("cosign is disabled")
		return false, nil
	}
	if !c.cosign.GetEnabled() {
		return false, nil
	}
	// Verify would pass without running cosign, and recording that as verified would
	// let an install skip a check that never happened.
	if !c.verifier.Supported() {
		logger.Debug("cosign doesn't support this environment")
		return false, nil
	}
	return true, nil
}

// VerifyDigest verifies the signature against the asset's digest, without the
// asset. It returns false when the signature can't be verified that way.
func (c *cosignVerifier) VerifyDigest(ctx context.Context, logger *slog.Logger, digest string) (bool, error) {
	logger.Info("verifying a digest with Cosign")
	if err := c.installer.install(ctx, logger); err != nil {
		return false, fmt.Errorf("install sigstore/cosign: %w", err)
	}
	f, err := c.verifier.VerifyDigest(ctx, logger, c.runtime, c.file(), c.cosign, c.pkg.TemplateArtifact(c.runtime, c.asset), digest)
	if err != nil {
		return false, fmt.Errorf("verify a digest with Cosign: %w", err)
	}
	return f, nil
}

func (c *cosignVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verifying a file with Cosign")
	if err := c.installer.install(ctx, logger); err != nil {
		return fmt.Errorf("install sigstore/cosign: %w", err)
	}

	art := c.pkg.TemplateArtifact(c.runtime, c.asset)

	if err := c.verifier.Verify(ctx, logger, c.runtime, c.file(), c.cosign, art, file); err != nil {
		return fmt.Errorf("verify a file with Cosign: %w", err)
	}
	return nil
}

func (c *cosignVerifier) file() *download.File {
	pkg := c.pkg
	return &download.File{
		RepoOwner: pkg.PackageInfo.RepoOwner,
		RepoName:  pkg.PackageInfo.RepoName,
		Host:      pkg.PackageInfo.GetHost(),
		Version:   pkg.Package.Version,
	}
}
