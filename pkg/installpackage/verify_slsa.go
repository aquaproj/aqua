package installpackage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
)

type slsaVerifier struct {
	disabled  bool
	pkg       *config.Package
	installer *DedicatedInstaller
	verifier  SLSAVerifier
	runtime   *runtime.Runtime
	asset     string
}

func (s *slsaVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if s.disabled {
		logger.Debug("slsa verification is disabled")
		return false, nil
	}
	return s.pkg.PackageInfo.SLSAProvenance.GetEnabled(), nil
}

func (s *slsaVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verify a package with slsa-verifier")
	installerPkg := s.installer.Pkg()
	logger = logger.With(
		"package_name", installerPkg.Package.Name,
		"package_version", installerPkg.Package.Version,
		"registry", installerPkg.Package.Registry,
	)
	if err := s.installer.install(ctx, logger); err != nil {
		return fmt.Errorf("install slsa-verifier: %w", err)
	}

	pkg := s.pkg
	pkgInfo := s.pkg.PackageInfo

	art := pkg.TemplateArtifact(s.runtime, s.asset)
	sourceTag := pkgInfo.SLSAProvenance.SourceTag
	if sourceTag == "" {
		sourceTag = pkg.Package.Version
	}

	if err := s.verifier.Verify(ctx, logger, s.runtime, pkgInfo.SLSAProvenance, art, &download.File{
		RepoOwner: pkgInfo.RepoOwner,
		RepoName:  pkgInfo.RepoName,
		Version:   pkg.Package.Version,
	}, &slsa.ParamVerify{
		SourceURI:    pkgInfo.SLSASourceURI(),
		SourceTag:    sourceTag,
		ArtifactPath: file,
	}); err != nil {
		return fmt.Errorf("verify a package with slsa-verifier: %w", err)
	}
	return nil
}
