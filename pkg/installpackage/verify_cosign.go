package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
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

func (c *cosignVerifier) Enabled(logE *logrus.Entry) (bool, error) {
	if c.disabled {
		logE.Debug("cosign is disabled")
		return false, nil
	}

	return c.cosign.GetEnabled(), nil
}

func (c *cosignVerifier) Verify(ctx context.Context, logE *logrus.Entry, file string) error {
	logE.Info("verifying a file with Cosign")
	if err := c.installer.install(ctx, logE); err != nil {
		return fmt.Errorf("install sigstore/cosign: %w", err)
	}

	pkg := c.pkg
	cos := c.cosign

	art := pkg.TemplateArtifact(c.runtime, c.asset)

	if err := c.verifier.Verify(ctx, logE, c.runtime, &download.File{
		RepoOwner: pkg.PackageInfo.RepoOwner,
		RepoName:  pkg.PackageInfo.RepoName,
		Version:   pkg.Package.Version,
	}, cos, art, file); err != nil {
		return fmt.Errorf("verify a file with Cosign: %w", err)
	}
	return nil
}
